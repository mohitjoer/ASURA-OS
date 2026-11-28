import QtQuick
import Quickshell
import "installer"

// Toren interactive GUI installer. A normal floating window matched by Hyprland
// window rules on its title ("Toren Linux Installer"), hosted at the shell root
// so qs.* imports resolve cleanly.
ShellRoot {
    FloatingWindow {
        id: win
        title: "Toren Linux Installer"

        readonly property int fitW: win.screen ? Math.min(960, win.screen.width - 24) : 960
        readonly property int fitH: win.screen ? Math.min(680, win.screen.height - 56) : 680
        readonly property bool cramped: win.fitW < 960 || win.fitH < 680
        minimumSize: Qt.size(Math.min(820, win.fitW), Math.min(560, win.fitH))
        maximumSize: win.cramped ? Qt.size(win.fitW, win.fitH) : Qt.size(16777215, 16777215)
        onClosed: Qt.quit()

        Installer {
            anchors.fill: parent
        }
    }
}
