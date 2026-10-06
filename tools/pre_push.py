#!/usr/bin/env python3
"""检查待推送提交的隔离副本，失败时阻止远端引用更新。"""
import argparse
import datetime
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile

from verify import stop_processes

ROOT = Path(__file__).resolve().parent.parent
HOOK = '.githooks'


def git(*args, check=True):
    """读取和维护当前仓库的验证 worktree。"""
    result = subprocess.run(['git', *args], cwd=ROOT, text=True, stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, check=check)
    return result.stdout.strip()


def install():
    """安装仓库维护的 hook，已有配置需要先由维护者处理。"""
    current = git('config', '--get', 'core.hooksPath', check=False)
    if current and current != HOOK:
        raise RuntimeError('已有 core.hooksPath=' + current + '，请先处理自定义 hook 配置')
    existing = Path(git('rev-parse', '--git-path', 'hooks/pre-push'))
    if not existing.is_absolute():
        existing = ROOT / existing
    if not current and existing.exists():
        raise RuntimeError('已有 pre-push hook: ' + str(existing))
    git('config', '--local', 'core.hooksPath', HOOK)
    print('已安装推送检查: ' + HOOK + '/pre-push')


def commit(oid, remote):
    """远端旧提交可能不在本地，获取后再解析 annotated tag。"""
    try:
        return git('rev-parse', '--verify', oid + '^{commit}')
    except subprocess.CalledProcessError:
        git('fetch', '--no-tags', '--', remote, oid)
        return git('rev-parse', '--verify', oid + '^{commit}')


def empty_tree():
    """写入空树对象，供没有共同祖先的引用比较使用。"""
    return subprocess.check_output(['git', 'hash-object', '-w', '-t', 'tree', '--stdin'],
                                   cwd=ROOT, input=b'').decode().strip()


def new_ref_base(sha, remote):
    """从远端默认分支计算基准，空远端使用空树。"""
    listing = git('ls-remote', '--symref', '--', remote, 'HEAD')
    lines = listing.splitlines()
    default_oid = next((line.split()[0] for line in lines if not line.startswith('ref:')), None)
    if not default_oid:
        if git('ls-remote', '--heads', '--', remote):
            raise RuntimeError('远端存在分支，但无法确定默认分支')
        return empty_tree()
    base = commit(default_oid, remote)
    result = subprocess.run(['git', 'merge-base', base, sha], cwd=ROOT, text=True,
                            stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if result.returncode == 1:
        return empty_tree()
    result.check_returncode()
    return result.stdout.strip()


def updates(stream, remote):
    """同一提交合并验证，保留各引用对应的格式比较区间。"""
    selected = {}
    for line in stream:
        fields = line.split()
        if len(fields) != 4:
            raise RuntimeError('Git 引用更新格式无效')
        local_ref, local_oid, remote_ref, remote_oid = fields
        if set(local_oid) == {'0'}:
            continue
        sha = git('rev-parse', '--verify', local_oid + '^{commit}')
        base = new_ref_base(sha, remote) if set(remote_oid) == {'0'} else commit(remote_oid, remote)
        entry = selected.setdefault(sha, {'bases': [], 'refs': []})
        if base not in entry['bases']:
            entry['bases'].append(base)
        entry['refs'].append(remote_ref)
    return selected


def validate(selected):
    """临时检出使用待推送内容，日志在 worktree 清理后继续保留。"""
    common = Path(git('rev-parse', '--git-common-dir'))
    if not common.is_absolute():
        common = ROOT / common
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ')
    for sha, entry in selected.items():
        logs = common / 'verification' / ('push-' + stamp) / sha
        logs.mkdir(parents=True, exist_ok=True)
        (logs / 'push.json').write_text(json.dumps({'commit': sha, **entry}, indent=2) + '\n')
        # 清除指向原检出的 Git 环境变量，临时副本自行解析 .git。
        env = os.environ.copy()
        for key in git('rev-parse', '--local-env-vars').splitlines():
            env.pop(key, None)
        env.update(VERIFY_BASES=json.dumps(entry['bases']), VERIFY_LOG_DIR=str(logs))
        with tempfile.TemporaryDirectory(prefix='tokenrouter-verify-') as temporary:
            tree = Path(temporary) / 'source'
            handlers = {}
            try:
                git('worktree', 'add', '--detach', str(tree), sha)
                print('验证待推送提交: ' + sha + '\n日志: ' + str(logs), flush=True)
                with (logs / 'push.log').open('w') as log:
                    process = subprocess.Popen(['make', 'verify'], cwd=tree, env=env,
                                               stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                               text=True, start_new_session=True)
                    try:
                        for line in process.stdout:
                            log.write(line)
                            log.flush()
                            print(line, end='', flush=True)
                        code = process.wait()
                    finally:
                        # make 退出后调度器可能还在收回独立进程组，等待整组退出。
                        handlers = {sig: signal.signal(sig, signal.SIG_IGN)
                                    for sig in (signal.SIGTERM, signal.SIGINT)}
                        stop_processes([process], timeout=10)
                if code:
                    raise RuntimeError('提交 ' + sha + ' 验证失败，日志: ' + str(logs))
            finally:
                try:
                    # 测试进程全部结束后移除构建生成物和临时检出。
                    if tree.exists():
                        git('worktree', 'remove', '--force', str(tree))
                finally:
                    for sig, handler in handlers.items():
                        signal.signal(sig, handler)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--install', action='store_true')
    parser.add_argument('remote', nargs='?')
    parser.add_argument('url', nargs='?')
    args = parser.parse_args()
    if args.install:
        install()
    elif args.remote and args.url:
        validate(updates(sys.stdin, args.url))
    else:
        parser.error('需要 Git 提供远端名称和地址')


if __name__ == '__main__':
    signal.signal(signal.SIGTERM, lambda *_: sys.exit(143))
    try:
        main()
    except (OSError, RuntimeError, subprocess.CalledProcessError, KeyboardInterrupt) as error:
        print('推送检查未通过: ' + str(error), file=sys.stderr)
        sys.exit(1)
