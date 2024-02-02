.PHONY:
git-sync:
	git pull
	git add .
	git commit -m "update"
	git push
