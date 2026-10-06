"""用隔离进程检查依赖、并发和取消后的资源释放。"""
import json
import os
import signal
import subprocess
import sys
import time

from test_verification import GitFixture


class ScheduleTest(GitFixture):
    """夹具使用事件文件控制执行顺序，避免依赖机器速度。"""

    def prepare(self, graph, jobs=2):
        (self.root / 'graph.json').write_text(json.dumps(graph))
        (self.root / 'runner.py').write_text(
            'import sys,json,signal\nsys.path.insert(0,"tools")\nimport verify\nsignal.signal(signal.SIGTERM, lambda *_: sys.exit(143))\n'
            f'raise SystemExit(verify.run_targets([],graph=json.load(open("graph.json")),jobs={jobs}))\n')
        (self.root / 'stage.py').write_text(
            'import sys,time,os,signal,subprocess\nfrom pathlib import Path\n'
            'name=sys.argv[1]\nPath(name+".started").write_text(str(time.monotonic()))\n'
            'if name=="left":\n'
            ' while not Path("right.started").exists(): time.sleep(.01)\n'
            'if name=="right":\n'
            ' while not Path("left.started").exists(): time.sleep(.01)\n'
            'if name=="fail":\n'
            ' while not Path("child.pid").exists(): time.sleep(.01)\n'
            ' raise SystemExit(2)\n'
            'if name=="long":\n'
            ' child=subprocess.Popen([sys.executable,"-c",'
            '"import time;time.sleep(60)"])\n'
            ' def stop(*args):\n'
            '  signal.signal(signal.SIGTERM,signal.SIG_IGN)\n'
            '  child.wait()\n'
            '  Path("cleaned").write_text("yes")\n'
            '  raise SystemExit(143)\n'
            ' signal.signal(signal.SIGTERM,stop)\n'
            ' Path("child.pid").write_text(str(child.pid))\n'
            ' time.sleep(60)\n'
            'Path(name+".finished").write_text(str(time.monotonic()))\n')
        (self.root / 'Makefile').write_text(''.join(
            f'{name}:\n\t@{sys.executable} stage.py {name}\n' for name in graph))

    def reports(self):
        paths = list((self.root / '.git/verification').glob('*/summary.json'))
        self.assertEqual(len(paths), 1)
        return json.loads(paths[0].read_text())

    def test_parallel_dependencies_and_timings(self):
        self.prepare({'first': [], 'left': ['first'], 'right': ['first'],
                      'third': ['first'], 'last': ['left', 'right', 'third']})
        result = self.command([sys.executable, 'runner.py'])
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        report = self.reports()
        self.assertTrue(all(c['status'] == '通过' and c['duration_seconds'] >= 0 for c in report['checks']))
        when = lambda name: float((self.root / name).read_text())
        self.assertLess(when('first.finished'), when('left.started'))
        self.assertLess(when('first.finished'), when('right.started'))
        self.assertGreaterEqual(when('third.started'), min(when('left.finished'), when('right.finished')))
        self.assertGreater(when('last.started'), max(when('left.finished'), when('right.finished'), when('third.finished')))

    def test_serial_mode(self):
        self.prepare({'a': [], 'b': [], 'c': []}, jobs=1)
        result = self.command([sys.executable, 'runner.py'])
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertLess(float((self.root / 'a.finished').read_text()), float((self.root / 'b.started').read_text()))
        self.assertLess(float((self.root / 'b.finished').read_text()), float((self.root / 'c.started').read_text()))

    def test_failure_cancels_children_and_blocks_dependents(self):
        self.prepare({'long': [], 'fail': [], 'later': ['fail']})
        result = self.command([sys.executable, 'runner.py'])
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual({c['target']: c['status'] for c in self.reports()['checks']},
                         {'long': '取消', 'fail': '失败', 'later': '未执行'})
        self.assertTrue((self.root / 'cleaned').exists())
        with self.assertRaises(ProcessLookupError):
            os.kill(int((self.root / 'child.pid').read_text()), 0)
        self.assertFalse((self.root / 'later.started').exists())

    def test_signal_cancels_children_before_exit(self):
        self.prepare({'long': [], 'later': ['long']})
        process = subprocess.Popen([sys.executable, 'runner.py'], cwd=self.root,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            deadline = time.monotonic() + 10
            while not (self.root / 'child.pid').exists():
                self.assertIsNone(process.poll())
                self.assertLess(time.monotonic(), deadline)
                time.sleep(.01)
            process.terminate()
            process.communicate(timeout=15)
            self.assertNotEqual(process.returncode, 0)
            self.assertTrue((self.root / 'cleaned').exists())
            with self.assertRaises(ProcessLookupError):
                os.kill(int((self.root / 'child.pid').read_text()), 0)
            self.assertEqual(self.reports()['checks'][0]['status'], '取消')
        finally:
            if process.poll() is None:
                process.kill()
                process.communicate()

    def test_invalid_concurrency_rejected(self):
        self.prepare({'a': []}, jobs=3)
        result = self.command([sys.executable, 'runner.py'])
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse((self.root / 'a.started').exists())

    def test_push_interrupt_waits_for_scheduled_processes(self):
        self.prepare({'long': []})
        stage = self.root / 'stage.py'
        # 检出被删除后仍保留清理证据，验证 hook 等待了调度器的子进程。
        stage.write_text(stage.read_text().replace(
            'Path("child.pid")', 'Path(os.environ["VERIFY_LOG_DIR"], "child.pid")').replace(
            'Path("cleaned")', 'Path(os.environ["VERIFY_LOG_DIR"], "cleaned")'))
        with (self.root / 'Makefile').open('a') as file:
            file.write(f'verify:\n\t@{sys.executable} runner.py\n')
        self.commit()
        process = subprocess.Popen(['git', 'push', 'origin', 'main'], cwd=self.root,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE, start_new_session=True)
        try:
            deadline = time.monotonic() + 15
            markers = []
            while not markers:
                self.assertIsNone(process.poll())
                self.assertLess(time.monotonic(), deadline)
                markers = list((self.root / '.git/verification').glob('push-*/*/child.pid'))
                time.sleep(.01)
            # 对 git 所在进程组中断，包含 hook；测试自身处于另一个组。
            os.killpg(process.pid, signal.SIGTERM)
            process.communicate(timeout=20)
            self.assertNotEqual(process.returncode, 0)
            self.assertTrue((markers[0].parent / 'cleaned').exists())
            with self.assertRaises(ProcessLookupError):
                os.kill(int(markers[0].read_text()), 0)
            self.assert_cleaned()
            self.assertEqual(self.git('ls-remote', 'origin', 'refs/heads/main'), '')
        finally:
            if process.poll() is None:
                process.kill()
                process.communicate()
