pragma Singleton
// ASURA overlay (shell-desktop/toren/vendor.sh OVERLAYS): the bar always stays in
// the focus grab, so opening a sidebar from it does not close it again.
pragma ComponentBehavior: Bound
import QtQuick
import Quickshell
import Quickshell.Hyprland

/**
 * Manages a HyprlandFocusGrab that's to be shared by all windows.
 * "Persistent" is for windows that should always be included but not closed on dismiss, like bar and onscreen keyboard.
 * "Dismissable" is for stuff like sidebars
 */ 
Singleton {
    id: root

    signal dismissed()

    property list<var> persistent: []
    property list<var> dismissable: []

    function dismiss() {
        root.dismissable = [];
        root.dismissed();
    }

    Component.onCompleted: {
        console.log("[GlobalFocusGrab] Initialized");
    }

    function addPersistent(window) {
        if (root.persistent.indexOf(window) === -1) {
            root.persistent.push(window);
        }
    }

    function removePersistent(window) {
        var index = root.persistent.indexOf(window);
        if (index !== -1) {
            root.persistent.splice(index, 1);
        }
    }

    function addDismissable(window) {
        if (root.dismissable.indexOf(window) === -1) {
            root.dismissable.push(window);
        }
    }

    function removeDismissable(window) {
        var index = root.dismissable.indexOf(window);
        if (index !== -1) {
            root.dismissable.splice(index, 1);
        }
    }

    function hasActive(element) {
        return element?.activeFocus || Array.from(
            element?.children
        ).some(
            (child) => hasActive(child)
        );
    }

    // ASURA: the grab is armed a moment after something opens, once its
    // window is on screen. Armed in the same instant the window was made
    // visible, Hyprland 0.49 can clear the grab at once, and end-4 then closed
    // the panel it had just opened (the sidebars, the wallpaper picker). A
    // grab that is still cleared within its first moments is re-armed rather
    // than taken as a click outside. Every clear is logged with its timing.
    property bool armed: false
    property double armedAt: 0
    readonly property bool wanted: root.dismissable.length > 0
    onWantedChanged: {
        if (root.wanted) {
            armTimer.restart();
        } else {
            armTimer.stop();
            root.armed = false;
        }
    }
    Timer {
        id: armTimer
        interval: 250
        onTriggered: {
            if (!root.wanted)
                return;
            root.armedAt = Date.now();
            root.armed = true;
        }
    }

    HyprlandFocusGrab {
        id: grab
        // The persistent windows (the bar, the on-screen keyboard) are always
        // in the grab, so the window list does not change when focus arrives.
        windows: [...root.dismissable, ...root.persistent]
        active: root.armed && root.wanted
        onCleared: () => {
            const after = Date.now() - root.armedAt;
            if (root.wanted && after < 600) {
                console.log("[GlobalFocusGrab] cleared " + after + " ms after arming; re-arming");
                root.armed = false;
                armTimer.restart();
                return;
            }
            console.log("[GlobalFocusGrab] cleared " + after + " ms after arming; dismissing");
            root.dismiss();
        }
    }

}
