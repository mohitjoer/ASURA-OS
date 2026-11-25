import QtQuick
import Quickshell
import Quickshell.Hyprland
import Quickshell.Io
import qs
import qs.modules.common
import qs.modules.common.functions

// The ASURA bridge. ASURA's keybindings (session/hyprland/binds.conf) do not
// talk to a shell directly: they run `asura ipc call asura <function>`, and
// `asura ipc` delivers that to whichever desktop shell is running. In ASURA's
// own shell the `asura` target is in shell.qml; when the desktop runs,
// this file answers the same twelve functions by driving the desktop's own
// surfaces, so every shortcut keeps working without one binding changing.
//
// The desktop opens most surfaces through GlobalStates flags, which are set here
// directly. The cheatsheet, the region selector's actions and the clipboard
// search live inside their modules and are reachable only through the desktop's
// own IPC targets, so those are called through `quickshell ipc` on this same
// config.
Item {
    id: bridge

    readonly property string configPath: FileUtils.trimFileProtocol(Quickshell.shellPath(""))

    function callSelf(target: string, fn: string): void {
        Quickshell.execDetached(["quickshell", "ipc", "--path", bridge.configPath, "call", target, fn]);
    }

    // ── First login on this desktop ─────────────────────────────────────────
    // Once the desktop config has loaded: seed matugen and a palette from the
    // user's ASURA wallpaper (asura/first-run.sh, which never overwrites), and
    // keep the bar on top and pin the dock with ASURA's apps. The dock is set once per user,
    // remembered by a marker, so a user who turns it off keeps it off.
    readonly property string stateDir: (Quickshell.env("XDG_STATE_HOME") || (Quickshell.env("HOME") + "/.local/state")) + "/asura"
    property bool firstRunDone: false
    function firstRun(): void {
        if (bridge.firstRunDone || !Config.ready)
            return;
        bridge.firstRunDone = true;
        Quickshell.execDetached(["sh", bridge.configPath + "/asura/first-run.sh"]);
        dockMarker.running = true;
    }
    Connections {
        target: Config
        function onReadyChanged(): void { bridge.firstRun(); }
    }
    Component.onCompleted: {
        // A slim bar, like Ubuntu's and GNOME's: standard is 30 px. Every
        // bar size the desktop derives (barHeight, the reserved zone) follows it.
        Appearance.sizes.baseBarHeight = 30;
        bridge.firstRun();
    }

    // The desktop rounding stays as set. Appearance.rounding is not
    // only corner radii: it also uses those values as margins and padding,
    // and rounding.full as "a circle". ASURA's square parts are the
    // windows (asura/hyprland.conf), the plain bar, no screen corners, and its
    // own bar widgets, which set their own radii.
    Process {
        id: dockMarker
        command: ["sh", "-c", "m=\"$1/desktop-layout-6\"; [ -e \"$m\" ] && exit 1; mkdir -p \"$1\" && touch \"$m\"", "sh", bridge.stateDir]
        onExited: (code) => {
            if (code !== 0)
                return;
            // The bar stays on top.
            Config.options.bar.vertical = false;
            // ASURA's square look: a plain rectangular bar, no
            // rounded screen corners, the digital desktop clock.
            Config.options.bar.cornerStyle = 2;
            Config.options.appearance.fakeScreenRounding = 0;
            Config.options.background.widgets.clock.enable = false;
            // A plain bar: no boxes behind its groups.
            Config.options.bar.borderless = true;
            // Volume and brightness are sliders in the control menu (the right
            // sidebar), not scrolling on the bar's corners.
            Config.options.sidebar.quickSliders.enable = true;
            // The left sidebar (AI chat, anime) is not part of ASURA
            // (panelFamilies/IllogicalImpulseFamily.qml); its features are off.
            Config.options.policies.ai = 0;
            Config.options.policies.weeb = 0;
            Config.options.dock.enable = true;
            // The island dock auto-hides (Settings > Interface > Dock).
            Config.options.dock.pinnedOnStartup = false;
            // Apps in their own colours.
            Config.options.dock.monochromeIcons = false;
            Config.options.dock.pinnedApps = ["com.mitchellh.ghostty", "thunar", "chromium", "asura-store"];
        }
    }

    IpcHandler {
        target: "asura"

        // Super+Space (launcher), Super+Alt+Space (menu): the search is
        // its overview. Super+Escape (quick) and Super+N (sidebar): the
        // quick settings and notifications are its right sidebar.
        function togglePopover(name: string): void {
            switch (name) {
            case "calendar":
                GlobalStates.calendarOpen = !GlobalStates.calendarOpen;
                break;
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

        // ASURA's desktops are blocks of workspaces: a
        // desktop step is a workspace step, the way the bar scrolls.
        function overviewDesktop(delta: int): void {
            Hyprland.dispatch(delta >= 0 ? "workspace r+1" : "workspace r-1");
        }

        // Alt+Tab. The compositor cycles.
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

        // ASURA surfaces with a desktop counterpart open it; the rest (stash,
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

        // ASURA's settings live in GSettings; the desktop reads its own config
        // file and watches it, so there is nothing to reload here.
        function reloadSettings(): void {
        }
    }
}
