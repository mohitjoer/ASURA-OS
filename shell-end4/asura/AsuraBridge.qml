import QtQuick
import Quickshell
import Quickshell.Hyprland
import Quickshell.Io
import qs
import qs.modules.common.functions

// The ASURA bridge. ASURA's keybindings (session/hyprland/binds.conf) do not
// talk to a shell directly: they run `asura ipc call asura <function>`, and
// `asura ipc` delivers that to whichever desktop shell is running. In ASURA's
// own shell the `asura` target is in shell.qml; when the end-4 desktop runs,
// this file answers the same twelve functions by driving end-4's own
// surfaces, so every shortcut keeps working without one binding changing.
//
// end-4 opens most surfaces through GlobalStates flags, which are set here
// directly. The cheatsheet, the region selector's actions and the clipboard
// search live inside their modules and are reachable only through end-4's
// own IPC targets, so those are called through `quickshell ipc` on this same
// config.
Item {
    id: bridge

    readonly property string configPath: FileUtils.trimFileProtocol(Quickshell.shellPath(""))

    function callSelf(target: string, fn: string): void {
        Quickshell.execDetached(["quickshell", "ipc", "--path", bridge.configPath, "call", target, fn]);
    }

    IpcHandler {
        target: "asura"

        // Super+Space (launcher), Super+Alt+Space (menu): end-4's search is
        // its overview. Super+Escape (quick) and Super+N (sidebar): end-4's
        // quick settings and notifications are its right sidebar.
        function togglePopover(name: string): void {
            switch (name) {
            case "quick":
            case "sidebar":
                GlobalStates.sidebarRightOpen = !GlobalStates.sidebarRightOpen;
                break;
            default:
                GlobalStates.overviewOpen = !GlobalStates.overviewOpen;
            }
        }

        function toggleOverview(): void {
            GlobalStates.overviewOpen = !GlobalStates.overviewOpen;
        }

        // ASURA's desktops are blocks of workspaces; end-4 has none, so a
        // desktop step is a workspace step, the way end-4's own bar scrolls.
        function overviewDesktop(delta: int): void {
            Hyprland.dispatch(delta >= 0 ? "workspace r+1" : "workspace r-1");
        }

        // Alt+Tab. end-4 has no window switcher, so the compositor cycles.
        function advanceSwitcher(): void {
            Hyprland.dispatch("cyclenext");
            Hyprland.dispatch("bringactivetotop");
        }
        function reverseSwitcher(): void {
            Hyprland.dispatch("cyclenext prev");
            Hyprland.dispatch("bringactivetotop");
        }
        function closeSwitcher(): void {
        }

        function toggleCheatsheet(): void {
            bridge.callSelf("cheatsheet", "toggle");
        }

        function session(): void {
            GlobalStates.sessionOpen = !GlobalStates.sessionOpen;
        }

        function ocr(): void {
            bridge.callSelf("region", "ocr");
        }

        function toggleScreenshot(): void {
            bridge.callSelf("region", "screenshot");
        }

        // ASURA surfaces with an end-4 counterpart open it; the rest (stash,
        // the visualizer) have none and do nothing in this desktop.
        function toggleSurface(name: string): void {
            switch (name) {
            case "overview":
                GlobalStates.overviewOpen = !GlobalStates.overviewOpen;
                break;
            case "clipboard":
                bridge.callSelf("search", "clipboardToggle");
                break;
            case "screenshot":
                bridge.callSelf("region", "screenshot");
                break;
            case "cheatsheet":
                bridge.callSelf("cheatsheet", "toggle");
                break;
            }
        }

        // ASURA's settings live in GSettings; end-4 reads its own config
        // file and watches it, so there is nothing to reload here.
        function reloadSettings(): void {
        }
    }
}
