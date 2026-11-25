# When something goes wrong

1. **Run `toren doctor`.** Every check, with what is wrong and often the fix.
2. **The desktop is blank or broken**: `tail -n 60 ~/.local/state/asura/quickshell.log`
   names the problem; `toren refresh` restarts the desktop.
3. **It broke after an update**: `sudo asura rollback --system NAME` with a
   snapshot from before it (`asura snapshot` lists them), then restart.
4. **Your settings are the problem**: `toren rollback` restores the previous
   ones; `toren reset PATH` puts one file back to Toren's version.
5. **No sound**: **Super + Shift + A** restarts audio.
6. **Ask**: open an issue at https://github.com/mohitjoer/Toren-Linux/issues with
   the output of `asura version -v` and `toren doctor --json`.
