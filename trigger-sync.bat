@echo off
echo ⚡ Triggering manual sync...
docker kill --signal=SIGUSR1 leetcode-sync
echo ✅ Signal sent! Check logs for progress.
