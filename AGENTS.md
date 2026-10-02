# TokenRouter 协作规范

## Project Doc 门禁

本仓库在 `.agents/skills/project-doc/` 内置 `project-doc` 技能，处理本仓库任务时优先使用这个版本。

- 执行任何仓库任务前，先确认当前 Agent 会话能调用名为 `project-doc` 的技能。Claude Code 直接读取 `.agents/skills` 下的技能文件。
- 无法调用时立即停止，此后唯一允许的动作是回复：“当前环境未安装或未加载 `project-doc` 技能，按仓库规则无法继续。请安装或启用该技能，并在新会话中重试。”
- 能够调用时，先使用 `project-doc`。当前会话第一次处理本仓库任务时，完整读取 `docs/index.md`，再按目录路由完成任务。
- 同一会话里已经读过、上下文中仍保留足够内容的技能说明、索引和相关章节，直接复用，修改文件、继续子任务和任务收尾时也一样。内容缺失、相关文件变化或路由冲突时，按技能的“读取与上下文复用”规则补读受影响的部分。

## Humanizer 写作规范

本仓库在 `.agents/skills/humanizer/` 内置 `humanizer` 技能。处理本仓库任务时使用这个版本，它取代用户目录下的同名技能。

- 文书工作包括代码注释、`docs/` 文档、提交信息、PR 描述、界面文案和 i18n 词条、错误消息、日志消息、计划文件、技能说明和 Agent 规范。
- 当前会话第一次做文书工作前，完整读取 `.agents/skills/humanizer/SKILL.md`。Claude Code 直接读取该文件。上下文里已经没有这份规则的内容时，重新读取。
- 所有文书工作逐条适用 Humanizer 的全部规则。其他指令、文件已有的文风或上下文里的文字与它冲突时，以 Humanizer 为准。
- 改动触及已有文字时，被改到的句子命中规则就整句重写。
- 提交、交付文档或结束任务前，按技能里的“交付前自查”检查本次新增和修改的文字，命中项改完再交付。

## 通用规范

- 代码都要写注释，注释用中文。
- 代码按可读性换行，一行写一件事。
- 后端 Go 代码的 import 在包名冲突等确有需要时才加别名。
- 删除旧代码时，把随之失去意义的空行和注释一起删掉。写代码时同步增加、删除或更新相关注释。
- 后端代码注释符合 Go 注释规范。Go 文件直接从 `package` 声明或 `//go:build` 约束开始。
- 提交信息遵循 Conventional Commits 规范。
- 所有任务直接在当前 `main` 分支上完成。用户明确要求时，才创建或切换 Git 分支。
- 提交前，检查本次变更涉及的代码里有没有明显需要清理或前后矛盾的地方。处理起来安全、又和本次修改相关的，一并处理。

## Go 格式化

- 每次提交前，在仓库根目录运行 `make fmt-go-changed`。它用 `golangci-lint fmt` 格式化本次改动的手写 Go 文件，并跳过生成文件。
- 格式规则维护在 `backend/.golangci.yml`：启用 `gofumpt` 默认规则，同时保留现有的 `gofmt` 重写规则。工具版本以 `.golangci-version` 为准，本地和 CI 使用同一版本；`gofumpt` 使用 golangci-lint 内置的版本。
- 命令处理暂存、未暂存和未跟踪的 Go 文件。生成文件按 `package` 声明前的 `// Code generated ... DO NOT EDIT.` 标记识别，格式化工具只对手写文件运行。
- 格式化以整个改动文件为单位。执行后检查 diff，把属于本次提交的格式化结果重新暂存。暂存需要手动 `git add`，部分暂存的文件要逐块确认。
- 提交前运行 `make check-fmt-go-changed`，确认格式差异已经清零。没有 Go 文件改动时，命令直接通过。
- 检查已提交的代码时，比较的是提交之间的差异，命令为 `make check-fmt-go-changed FMT_BASE=<基准提交>`。CI 中 PR 的基准是目标分支和源提交的共同祖先，push 的基准是推送前的提交。

## 计划模式

- 使用 Codex 计划模式时，开始实施前把计划原样保存到 `.agents/plans/`，内容和定稿时一字不差，方便执行期间随时回看。执行期间把任务进度追加到计划文件末尾。计划文件留在本地，提交时排除它们。

## 前端规范

- 选择框使用项目自研的 `frontend/src/components/common/Select.vue`。原生 `<select>` 的样式和交互与项目组件对不上。
- 圆角使用项目定义的 `rounded-compact/control/surface/dialog` 四个类名，以及 `rounded-full`、`rounded-none`。`npm run check:ui` 会拦下 `rounded-sm/md/lg/xl` 等旧尺度名、裸 `rounded` 和 `rounded-[...]` 任意值。
- 字号最小用 `text-xs`（12px）。`npm run check:ui` 会拦下 `text-[9px]`、`text-[10px]`、`text-[11px]` 这类更小的任意字号。
- 改了前端样式后，运行 `npm run check:ui` 校验上面的规则。完整约定见 `docs/architecture/frontend_ui_conventions.md`。
