chrome-profile-router: *.go *.h *.m Makefile
	go build -o ChromeProfileRouter.app/Contents/MacOS/chrome-profile-router

.PHONY: clean
clean:
	rm -f ChromeProfileRouter.app/Contents/MacOS/chrome-profile-router

unlock:
	killall -9 chrome-profile-router || true
	rm "${TMPDIR}chrome-profile-router.lock" || true
