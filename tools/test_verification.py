"""用隔离仓库验证推送拦截、提交范围和检查结果。"""
import contextlib
import io
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import tempfile
import time
import unittest
from unittest.mock import patch

import verify
from check_pnpm_audit_exceptions import main as audit_main, validate_audit

SOURCE = Path(__file__).resolve().parent.parent


def audit_report(advisories=None):
    """构造含统计信息的 pnpm 审计报告。"""
    advisories = advisories or {}
    counts = dict.fromkeys(('info', 'low', 'moderate', 'high', 'critical'), 0)
    for advisory in advisories.values():
        counts[advisory['severity']] += 1
    return {'advisories': advisories, 'metadata': {'vulnerabilities': counts}}


class AuditValidationTest(unittest.TestCase):
    """扫描失败与已知漏洞分别影响验证结果。"""

    def test_valid_empty_report(self):
        validate_audit(audit_report(), 0)

    def test_rejects_failed_or_incomplete_reports(self):
        for report in ({}, {'error': {'code': 'UNAVAILABLE'}}, {'advisories': {}}, [],
                       audit_report({'a': {'severity': 'high'}})):
            with self.subTest(report=report), self.assertRaises((ValueError, TypeError)):
                validate_audit(report, 0)
        with self.assertRaises(ValueError):
            validate_audit(audit_report(), 1)
        with self.assertRaises(ValueError):
            validate_audit(audit_report(), 2)

    def test_high_vulnerability_exceptions_and_expiry(self):
        report = audit_report({'a': {'module_name': 'example', 'severity': 'high',
                                     'github_advisory_id': 'GHSA-example'}})
        for expiry, expected in (('2999-01-01', 0), ('2000-01-01', 1), (None, 1)):
            exceptions = 'version: 1\nexceptions:\n'
            if expiry:
                exceptions += ('  - package: example\n    advisory: GHSA-example\n    severity: high\n'
                               '    mitigation: verified fixture\n    expires_on: ' + expiry + '\n')
            def fixture(path, *args, **kwargs):
                return io.StringIO(json.dumps(report) if path == 'audit.json' else exceptions)
            with self.subTest(expiry=expiry), patch('builtins.open', fixture), patch('sys.argv',
                    ['audit', '--audit', 'audit.json', '--exit-code', '1', '--exceptions', 'exceptions.yml']), \
                    contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
                self.assertEqual(audit_main(), expected)

    def test_missing_advisory_and_inconsistent_counts(self):
        report = audit_report({'a': {'module_name': 'example', 'severity': 'high'}})
        with self.assertRaises(ValueError):
            validate_audit(report)
        report = audit_report()
        report['metadata']['vulnerabilities']['critical'] = 1
        with self.assertRaises(ValueError):
            validate_audit(report)
        report = audit_report({'a': {'module_name': 'example', 'severity': 'high',
                                     'github_advisory_id': 'GHSA-example'}})
        with self.assertRaises(ValueError):
            validate_audit(report, 0)


class GitFixture(unittest.TestCase):
    """所有提交、推送和 hook 配置都局限在临时仓库。"""

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.directory = Path(self.temporary.name)
        self.root = self.directory / 'source'
        self.remote = self.directory / 'remote.git'
        subprocess.run(['git', 'init', '--bare', '--initial-branch=main', str(self.remote)],
                       check=True, capture_output=True)
        self.root.mkdir()
        self.git('init', '--initial-branch=main')
        self.git('config', 'user.name', 'Verification Test')
        self.git('config', 'user.email', 'verification@example.invalid')
        self.git('config', 'commit.gpgsign', 'false')
        self.git('remote', 'add', 'origin', str(self.remote))
        for name in ('tools/pre_push.py', 'tools/verify.py', '.githooks/pre-push'):
            target = self.root / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(SOURCE / name, target)
        (self.root / 'Makefile').write_text('verify:\n\t@python3 verify_case.py\n')
        # 夹具从提交内容决定通过或失败，并留下明确的启动标记。
        (self.root / 'verify_case.py').write_text(
            'import os,json,time\nfrom pathlib import Path\n'
            'logs=Path(os.environ["VERIFY_LOG_DIR"])\n'
            '(logs/"started").write_text("yes")\n'
            '(logs/"bases.json").write_text(os.environ["VERIFY_BASES"])\n'
            'if Path("sleep").exists(): time.sleep(60)\n'
            'raise SystemExit(1 if Path("broken").exists() else 0)\n')
        self.commit()
        result = self.command(['python3', 'tools/pre_push.py', '--install'])
        self.assertEqual(result.returncode, 0, result.stderr)

    def command(self, command):
        return subprocess.run(command, cwd=self.root, text=True, capture_output=True)

    def git(self, *args):
        result = self.command(['git', *args])
        if result.returncode:
            raise AssertionError(result.stderr)
        return result.stdout.strip()

    def commit(self):
        self.git('add', '.')
        self.git('commit', '-qm', 'test: verification fixture')
        return self.git('rev-parse', 'HEAD')

    def push(self, *refs):
        return self.command(['git', 'push', 'origin', *(refs or ('main',))])

    def assert_cleaned(self):
        self.assertEqual(self.git('worktree', 'list', '--porcelain').count('worktree '), 1)


class PrePushTest(GitFixture):
    """通过本地 bare 远端确认失败确实阻止引用更新。"""

    def test_committed_failure_cannot_be_hidden_by_working_tree(self):
        (self.root / 'broken').write_text('unit compilation failure')
        broken_sha = self.commit()
        (self.root / 'broken').unlink()
        (self.root / 'untracked.txt').write_text('keep')
        before = self.git('status', '--porcelain')
        result = self.push()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(broken_sha, result.stderr)
        self.assertEqual(self.git('ls-remote', 'origin', 'refs/heads/main'), '')
        self.assertEqual(self.git('status', '--porcelain'), before)
        self.assert_cleaned()
        self.commit()
        result = self.push()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn(self.git('rev-parse', 'HEAD'), self.git('ls-remote', 'origin', 'refs/heads/main'))
        self.assert_cleaned()

    def test_unit_compile_failure_blocks_remote_update(self):
        (self.root / 'go.mod').write_text('module fixture\n\ngo 1.27.0\n')
        (self.root / 'fixture.go').write_text('package fixture\n')
        (self.root / 'fixture_test.go').write_text('//go:build unit\n\npackage fixture\nimport "strings"\n')
        (self.root / 'Makefile').write_text('verify:\n\t@go test -tags=unit ./...\n')
        self.commit()
        result = self.push()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('imported and not used', result.stdout + result.stderr)
        self.assertEqual(self.git('ls-remote', 'origin', 'refs/heads/main'), '')
        self.assert_cleaned()

    def test_multiple_refs_annotated_tag_and_delete(self):
        self.git('branch', 'another')
        self.git('tag', '-am', 'fixture', 'v1.0.0')
        result = self.push('main', 'another', 'v1.0.0')
        self.assertEqual(result.returncode, 0, result.stderr)
        logs = list((self.root / '.git/verification').glob('push-*/*/push.json'))
        self.assertEqual(len(logs), 1)
        self.assertEqual(len(json.loads(logs[0].read_text())['refs']), 3)
        result = self.push(':another')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(len(list((self.root / '.git/verification').glob('push-*/*/push.json'))), 1)
        self.assert_cleaned()

    def test_existing_remote_base_covers_all_unpushed_commits(self):
        self.assertEqual(self.push().returncode, 0)
        base = self.git('rev-parse', 'HEAD')
        for name in ('first.go', 'second.go'):
            (self.root / name).write_text('package fixture\n')
            self.commit()
        self.assertEqual(self.push().returncode, 0)
        reports = list((self.root / '.git/verification').glob('push-*/*/push.json'))
        report = next(json.loads(p.read_text()) for p in reports
                      if json.loads(p.read_text())['commit'] == self.git('rev-parse', 'HEAD'))
        self.assertEqual(report['bases'], [base])

    def test_installer_is_idempotent_and_rejects_custom_hook(self):
        self.assertEqual(self.command(['python3', 'tools/pre_push.py', '--install']).returncode, 0)
        self.git('config', 'core.hooksPath', '.custom-hooks')
        self.assertNotEqual(self.command(['python3', 'tools/pre_push.py', '--install']).returncode, 0)
        self.assertEqual(self.git('config', '--get', 'core.hooksPath'), '.custom-hooks')

    def test_interrupt_cleans_worktree_and_blocks_push(self):
        (self.root / 'sleep').write_text('wait')
        self.commit()
        process = subprocess.Popen(['git', 'push', 'origin', 'main'], cwd=self.root,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                   start_new_session=True)
        try:
            deadline = time.monotonic() + 30
            while not list((self.root / '.git/verification').glob('push-*/*/started')):
                if process.poll() is not None:
                    self.fail('推送在验证启动前退出: ' + str(process.returncode))
                if time.monotonic() > deadline:
                    self.fail('推送检查没有启动')
                time.sleep(0.05)
            os.killpg(process.pid, signal.SIGTERM)
            # hook 继承输出管道；读到 EOF 时，hook 已退出并完成 worktree 清理。
            # Git 主进程可能先退出，清理期间查询 worktree 会与目录删除竞争。
            process.communicate(timeout=15)
            self.assertNotEqual(process.returncode, 0)
            self.assert_cleaned()
            self.assertEqual(self.git('ls-remote', 'origin', 'refs/heads/main'), '')
        finally:
            if process.poll() is None:
                os.killpg(process.pid, signal.SIGKILL)
                process.communicate()


class VerificationRunnerTest(GitFixture):
    """任一检查失败都会阻止后续步骤，并留下未执行状态。"""

    def test_stage_failures_are_reported(self):
        for stage in ('unit-compile', 'tagged-lint', 'frontend', 'integration', 'embed-resource'):
            with self.subTest(stage=stage):
                (self.root / 'Makefile').write_text(
                    f'pass:\n\t@true\n{stage}:\n\t@echo fixture failure\n\t@false\nlater:\n\t@touch should-not-exist\n')
                result = self.command(['python3', 'tools/verify.py', 'run', 'pass', stage, 'later'])
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('未执行: later', result.stdout)
                self.assertIn('含未提交修改', result.stdout)
                self.assertFalse((self.root / 'should-not-exist').exists())

    def test_validation_excludes_live_provider_credentials(self):
        variables = {'OPENAI_API_KEY': 'fixture', 'QODER_RUN_REAL_API_TESTS': '1',
                     'QODER_RUN_LOCAL_AUTH_TESTS': '1'}
        with patch.dict(os.environ, variables):
            environment = verify.test_environment()
            for name, value in variables.items():
                self.assertNotIn(name, environment)
                self.assertEqual(os.environ[name], value)
            self.assertEqual(environment['TOKENROUTER_VERIFY_STRICT'], '1')

    def test_environment_mismatches_fail_before_tests(self):
        def fixture(command, cwd=None):
            values = {'go': 'go version go' + (SOURCE / 'backend/go.mod').read_text().split('\ngo ')[1].splitlines()[0] + ' darwin/arm64', 'docker': '29.5.2',
                      'pg_dump': 'pg_dump (PostgreSQL) 18.3', 'psql': 'psql (PostgreSQL) 18.3'}
            return values[command[0]]
        with patch.object(verify, 'output', side_effect=fixture), contextlib.redirect_stdout(io.StringIO()):
            verify.preflight('backend')
        for failed in ('go', 'docker', 'pg_dump', 'psql'):
            def broken(command, cwd=None):
                if command[0] == failed:
                    raise FileNotFoundError(failed)
                return fixture(command)
            with self.subTest(tool=failed), patch.object(verify, 'output', side_effect=broken), \
                    contextlib.redirect_stdout(io.StringIO()), self.assertRaises(FileNotFoundError):
                verify.preflight('backend')
        with patch.object(verify, 'output', return_value='go version go1.26.0 darwin/arm64'), \
                contextlib.redirect_stdout(io.StringIO()), self.assertRaises(RuntimeError):
            verify.preflight('go')


if __name__ == '__main__':
    unittest.main()
