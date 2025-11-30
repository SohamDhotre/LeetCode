# 🐳 LeetCode → GitHub Sync (Docker Setup)

Automated LeetCode to GitHub sync system running in Docker - **no local dependencies required!**

## 🚀 Quick Start (Docker - Recommended)

### Prerequisites
- **Docker Desktop** installed ([Download](https://www.docker.com/products/docker-desktop))
- **Git** configured with GitHub access
- **LeetCode Account** with solved problems

### Step 1: Get Your LeetCode Session Cookie

1. Log in to [LeetCode](https://leetcode.com)
2. Press `F12` to open Developer Tools
3. Go to **Application** tab (Chrome/Edge) or **Storage** tab (Firefox)
4. Navigate to **Cookies** → `https://leetcode.com`
5. Find `LEETCODE_SESSION` and copy its value

### Step 2: Configure Environment

```bash
# Copy the example environment file
copy .env.example .env

# Edit .env and add your session cookie
# LEETCODE_SESSION=your_actual_cookie_value_here
```

### Step 3: Initialize Git Repository

```bash
cd d:\LeetCode
git init
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
git config user.name "Your Name"
git config user.email "your.email@example.com"
```

### Step 4: Run the Sync

**Windows:**
```bash
.\run-docker.bat
```

**Linux/Mac:**
```bash
chmod +x run-docker.sh
./run-docker.sh
```

That's it! The Docker container will:
- ✅ Build the Go application automatically
- ✅ Fetch your LeetCode submissions
- ✅ Organize and create files
- ✅ Generate READMEs
- ✅ Commit and push to GitHub

## 📁 What You'll Get

```
LeetCode/
├── Arrays/
│   ├── Easy/
│   │   ├── 1.two-sum.java
│   │   └── 1.two-sum.md
│   └── Medium/
├── Dynamic Programming/
│   └── Hard/
├── README.md                    # Auto-generated index
├── .leetcode_sync_db.json      # Sync tracking
├── Dockerfile                   # Docker configuration
├── docker-compose.yml          # Docker Compose setup
└── run-docker.bat              # Easy run script
```

## 🔄 Automation Options

### Option 1: Windows Task Scheduler

1. Open **Task Scheduler**
2. Create Basic Task: "LeetCode Sync"
3. Trigger: Daily at preferred time
4. Action: Start a program
   - Program: `d:\LeetCode\run-docker.bat`
   - Start in: `d:\LeetCode`

### Option 2: Docker Compose with Cron (Linux/Mac)

Add to your crontab:
```bash
# Run daily at 11 PM
0 23 * * * cd /path/to/LeetCode && docker-compose up
```

### Option 3: GitHub Actions (Cloud Automation)

Create `.github/workflows/sync.yml` in your repo for cloud-based automation.

## 🛠️ Advanced Usage

### Manual Docker Commands

Build the image:
```bash
docker build -t leetcode-sync .
```

Run manually:
```bash
docker run --rm \
  -v ${PWD}:/workspace \
  -v ${HOME}/.gitconfig:/root/.gitconfig:ro \
  --env-file .env \
  leetcode-sync
```

### Debug Mode

Enable detailed logging:
```bash
# In .env file
DEBUG=true
```

### Custom Configuration

Edit `config.json` to customize:
- Category mappings
- Language extensions
- Commit message templates

## 🔧 Troubleshooting

### "LEETCODE_SESSION not set"
- Ensure `.env` file exists (not `.env.example`)
- Verify the session cookie is correctly pasted

### "Permission denied" (Git)
- Check GitHub authentication
- For HTTPS: May need Personal Access Token
- For SSH: Ensure SSH keys are mounted in docker-compose.yml

### Docker build fails
- Ensure Docker Desktop is running
- Check internet connection for downloading dependencies

### No submissions found
- Verify you have accepted submissions on LeetCode
- Increase `MAX_SUBMISSIONS_TO_CHECK` in `.env`

## 📊 Features

- ✅ **Zero local dependencies** - Everything runs in Docker
- ✅ **Automatic organization** - Category/Difficulty structure
- ✅ **Rich metadata** - Problem details, tags, difficulty
- ✅ **Multi-language support** - Java, Python, C++, Go, SQL, etc.
- ✅ **Deduplication** - Never sync the same problem twice
- ✅ **Auto README generation** - Both problem-level and master index
- ✅ **Git automation** - Automatic commit and push
- ✅ **Configurable** - Customize via config.json

## 🔐 Security Notes

- Your `.env` file is gitignored automatically
- Session cookies are kept local
- No credentials are stored in the Docker image

## 📝 Alternative: Local Build (Without Docker)

If you prefer to build locally, see [setup.md](setup.md) for instructions.

---

**Happy Coding! 🚀**

*Sync your LeetCode journey to GitHub with zero effort!*
