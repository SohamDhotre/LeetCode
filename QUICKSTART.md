# 🚀 LeetCode → GitHub Sync - Quick Start

**Automated LeetCode to GitHub sync with ZERO local dependencies!**

---

## ⚡ 3-Minute Setup

### 1️⃣ Install Docker Desktop
Download: https://www.docker.com/products/docker-desktop

### 2️⃣ Get LeetCode Session Cookie
1. Login to LeetCode
2. Press `F12` → **Application** → **Cookies** → `leetcode.com`
3. Copy `LEETCODE_SESSION` value

### 3️⃣ Configure
```bash
cd d:\LeetCode
copy .env.example .env
# Edit .env and paste your session cookie
```

### 4️⃣ Setup Git
```bash
git init
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
git config user.name "Your Name"
git config user.email "your@email.com"
```

### 5️⃣ Run!
```bash
.\run-docker.bat
```

**Done!** Your solutions are now on GitHub! 🎉

---

## 📖 Full Documentation

- **Docker Setup**: [README.docker.md](README.docker.md) ← **START HERE**
- **Local Build**: [setup.md](setup.md)
- **Implementation Details**: See artifacts folder
- **Future Ideas**: [FUTURE_ENHANCEMENTS.md](FUTURE_ENHANCEMENTS.md)

---

## 🔄 Daily Use

After solving problems on LeetCode:
```bash
.\run-docker.bat
```

That's it! Everything else is automatic.

---

## 🤖 Automation

**Windows Task Scheduler:**
1. Open Task Scheduler
2. Create task: Run `d:\LeetCode\run-docker.bat` daily
3. Never think about it again!

---

## 🆘 Troubleshooting

| Issue | Solution |
|-------|----------|
| "LEETCODE_SESSION not set" | Create `.env` from `.env.example` |
| "Docker not found" | Install Docker Desktop |
| "Permission denied" (git) | Setup GitHub authentication |
| No submissions found | Increase `MAX_SUBMISSIONS_TO_CHECK` in `.env` |

---

## 📁 What You Get

```
LeetCode/
├── Arrays/Easy/1.two-sum.java
├── DP/Hard/72.edit-distance.cpp
├── README.md  ← Auto-generated stats!
└── ...
```

---

## ✨ Features

✅ Zero manual work  
✅ Professional organization  
✅ Auto README generation  
✅ Multi-language support  
✅ No local dependencies  
✅ Fully automated  

---

**Questions?** Check [README.docker.md](README.docker.md) for detailed guide!
