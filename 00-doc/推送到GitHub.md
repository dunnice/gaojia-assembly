# 推送到 GitHub

## 1. 在 GitHub 上新建仓库

1. 打开 [https://github.com/new](https://github.com/new)
2. 仓库名建议：`gaojia-assembly`
3. 选择 **Private** 或 **Public**
4. **不要**勾选「Add a README file」等初始化选项（本地已有内容）
5. 点击 **Create repository**

## 2. 添加远程并推送

创建完成后，在终端执行（将 `YOUR_USERNAME` 替换为你的 GitHub 用户名）：

```bash
cd /Users/ice7/Documents/temp/gaojia-assembly
git remote add origin https://github.com/YOUR_USERNAME/gaojia-assembly.git
git branch -M main
git push -u origin main
```

若使用 SSH：

```bash
git remote add origin git@github.com:YOUR_USERNAME/gaojia-assembly.git
git push -u origin main
```

## 3. 当前状态

- 已完成：`git init`、`.gitignore`、首次提交
- 已忽略：`node_modules/`、`*.db`、`crawler/config.json`、`backend/target/` 等
