#!/usr/bin/env python3
"""执行共用检查，记录结果并拒绝未完成的必要验证。"""
import argparse
import datetime
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent


def output(command, cwd=ROOT):
    """读取命令结果，非零退出码由调用者处理。"""
    return subprocess.check_output(command, cwd=cwd, text=True, stderr=subprocess.STDOUT).strip()


def git(*args):
    """在当前检出中执行 Git，兼容临时 worktree。"""
    return output(['git', *args])


def frontend_command(*args):
    """由同一组版本声明启动 Node 和 pnpm，复用 npm 的下载缓存。"""
    return ['npx', '--yes', '--package=node@' + (ROOT / '.node-version').read_text().strip(),
            '--package=pnpm@' + (ROOT / '.pnpm-version').read_text().strip(), '--', *args]


def preflight(kind):
    """在测试启动前验证工具和服务，失败时直接返回错误。"""
    if sys.version_info < (3, 10):
        raise RuntimeError('验证工具需要 Python 3.10 或更高版本')
    if kind in ('all', 'go', 'backend'):
        expected = re.search(r'^go (\S+)', (ROOT / 'backend/go.mod').read_text(), re.M)[1]
        actual = output(['go', 'version'])
        print(actual, flush=True)
        if actual.split()[2] != 'go' + expected:
            raise RuntimeError('Go 版本需要 ' + expected)
    if kind in ('all', 'frontend'):
        actual = output(frontend_command('node', '--version'))
        print('Node ' + actual, flush=True)
        if actual.lstrip('v') != (ROOT / '.node-version').read_text().strip():
            raise RuntimeError('Node 版本与 .node-version 不符')
        actual = output(frontend_command('pnpm', '--version'))
        print('pnpm ' + actual, flush=True)
        if actual != (ROOT / '.pnpm-version').read_text().strip():
            raise RuntimeError('pnpm 版本与 .pnpm-version 不符')
    if kind == 'all':
        expected = (ROOT / '.golangci-version').read_text().strip().lstrip('v')
        actual = output(['golangci-lint', 'version', '--short'])
        print('golangci-lint ' + actual, flush=True)
        if actual.lstrip('v') != expected:
            raise RuntimeError('golangci-lint 版本需要 ' + expected)
    if kind in ('all', 'backend'):
        print('Docker ' + output(['docker', 'info', '--format', '{{.ServerVersion}}']), flush=True)
        for tool in ('pg_dump', 'psql'):
            actual = output([tool, '--version'])
            print(actual, flush=True)
            if not re.search(r'\b18\.', actual):
                raise RuntimeError(tool + ' 需要 PostgreSQL 18 客户端')


def verification_bases():
    """推送使用每个引用的旧提交；手动检查使用上游共同祖先。"""
    if os.environ.get('VERIFY_BASES'):
        return json.loads(os.environ['VERIFY_BASES'])
    if os.environ.get('VERIFY_BASE'):
        return [os.environ['VERIFY_BASE']]
    try:
        upstream = git('rev-parse', '--verify', '@{upstream}')
    except subprocess.CalledProcessError:
        return [subprocess.check_output(['git', 'hash-object', '-w', '-t', 'tree', '--stdin'], cwd=ROOT, input=b'').decode().strip()]
    return [git('merge-base', upstream, 'HEAD')]


def check_diff():
    """同时检查提交区间和工作区；格式化命令使用检查模式。"""
    for base in verification_bases():
        subprocess.run(['git', 'diff', '--check', base, 'HEAD', '--'], cwd=ROOT, check=True)
        subprocess.run([sys.executable, 'tools/format_go.py', '--check', '--base', base], cwd=ROOT, check=True)
    for args in (['diff', '--check'], ['diff', '--cached', '--check']):
        subprocess.run(['git', *args], cwd=ROOT, check=True)
    subprocess.run([sys.executable, 'tools/format_go.py', '--check'], cwd=ROOT, check=True)


def test_environment():
    """供应商实测单独执行，完整验证不读取本机凭据。"""
    env = dict(os.environ, CI='true', TOKENROUTER_VERIFY_STRICT='1')
    for key in ('OPENAI_API_KEY', 'QODER_RUN_REAL_API_TESTS', 'QODER_RUN_LOCAL_AUTH_TESTS'):
        env.pop(key, None)
    return env


def go_test(tag):
    """保存 Go JSON 事件，汇总跳过项；必需环境缺失不能算通过。"""
    command = ['go', 'test', '-json', '-count=1', '-timeout=20m']
    if tag != 'ordinary':
        command += ['-tags=' + tag]
    if tag == 'integration':
        command += ['-p=4']
    command += ['./...']
    env = test_environment()
    process = subprocess.Popen(command, cwd=ROOT / 'backend', env=env, stdout=subprocess.PIPE,
                               stderr=subprocess.STDOUT, text=True)
    skipped = []
    counts = {"pass": 0, "fail": 0}
    details = {}
    try:
        for line in process.stdout:
            print(line, end='', flush=True)
            try:
                event = json.loads(line)
            except ValueError:
                continue
            key = (event.get('Package', ''), event.get('Test', ''))
            if event.get('Action') == 'output':
                details.setdefault(key, []).append(event.get('Output', ''))
            if event.get('Test') and event.get('Action') in counts:
                counts[event['Action']] += 1
            if event.get('Action') == 'skip' and event.get('Test'):
                skipped.append(key)
        code = process.wait()
    finally:
        if process.poll() is None:
            process.terminate()
            process.wait()
    for key in skipped:
        reason = ''.join(details.get(key, []))
        print('未执行: ' + '/'.join(key) + '\n' + reason, flush=True)
        if re.search(r'Docker|docker|pg_dump|psql|local listeners are not permitted', reason):
            code = 1
    print(f"Go {tag}: {counts['pass']} 通过，{counts['fail']} 失败，{len(skipped)} 未执行", flush=True)
    if code == 0 and counts['pass'] == 0:
        print('没有测试行为通过，验证未完成', file=sys.stderr)
        return 1
    return code


def run_targets(targets):
    """逐组运行 Make 目标，将日志和最终状态写入 Git 元数据目录。"""
    common = Path(git('rev-parse', '--git-common-dir'))
    if not common.is_absolute():
        common = ROOT / common
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ')
    directory = Path(os.environ.get('VERIFY_LOG_DIR', common / 'verification' / stamp))
    directory.mkdir(parents=True, exist_ok=True)
    report = {'commit': git('rev-parse', 'HEAD'), 'worktree_dirty': bool(git('status', '--porcelain')),
              'log_dir': str(directory),
              'checks': [{'target': target, 'status': '未执行'} for target in targets]}
    state = '含未提交修改' if report['worktree_dirty'] else '干净'
    print('检查工作区: HEAD=' + report['commit'] + '（' + state + '）\n日志目录: ' + str(directory), flush=True)
    code = 1
    try:
        for check in report['checks']:
            target = check['target']
            print('开始: make ' + target, flush=True)
            with (directory / (target + '.log')).open('w') as log:
                process = subprocess.Popen(['make', target], cwd=ROOT, stdout=log, stderr=subprocess.STDOUT)
                try:
                    code = process.wait()
                finally:
                    if process.poll() is None:
                        process.terminate()
                        process.wait()
            check['exit_code'] = code
            check['status'] = '通过' if code == 0 else '失败'
            print(check['status'] + ': make ' + target, flush=True)
            if code:
                print((directory / (target + '.log')).read_text(errors='replace')[-12000:], flush=True)
                break
    finally:
        (directory / 'summary.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
        for check in report['checks']:
            print(check['status'] + ': ' + check['target'], flush=True)
    return code


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action', choices=['preflight', 'diff', 'go-test', 'run', 'frontend'])
    parser.add_argument('arguments', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    if args.action == 'preflight':
        preflight(args.arguments[0])
    elif args.action == 'diff':
        check_diff()
    elif args.action == 'go-test':
        return go_test(args.arguments[0])
    elif args.action == 'run':
        return run_targets(args.arguments)
    elif args.action == 'frontend':
        return subprocess.call(frontend_command(*args.arguments), cwd=ROOT)
    return 0


if __name__ == '__main__':
    signal.signal(signal.SIGTERM, lambda *_: sys.exit(143))
    try:
        sys.exit(main())
    except (OSError, RuntimeError, subprocess.CalledProcessError, KeyboardInterrupt) as error:
        print('验证未完成: ' + str(error), file=sys.stderr)
        sys.exit(1)
