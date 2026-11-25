import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Wayland

// ASURA overlay (shell-desktop/toren/vendor.sh OVERLAYS): the desktop's dock as an
// island. It holds only the apps (the desktop's DockApps: pinned apps and running
// windows, with their previews), in a compact pill that floats above the
// bottom edge.
//
// Auto-hide (Settings > Interface > Dock, on by default): while a window is
// focused the island slides away and windows use the whole screen; it comes
// back when the pointer reaches the bottom edge, while a window preview is
// open, and on an empty workspace. With auto-hide off it is always shown and
// windows tile above it. The setting is the desktop's dock.pinnedOnStartup, read as
// "keep the dock shown".
//
// Removed from the desktop's dock: the pin button, the overview (apps grid) button
// and their separators. The geometry and the ids dockVisualBackground (with
// its margin) and dockRow (with its padding) are the desktop's: DockAppButton.qml
// reaches them by id.
Scope {
    id: root

    // The island: lifted this far off the edge, with these rounded ends.
    readonly property int islandLift: 10
    readonly property int islandRadius: 18
    readonly property int islandPadding: 6
    // The strip at the bottom edge that brings a hidden dock back.
    readonly property int revealStrip: 3
    readonly property bool autoHide: !(Config.options?.dock.pinnedOnStartup ?? false)

    Variants {
        // For each monitor
        model: Quickshell.screens

        PanelWindow {
            id: dockRoot
            required property var modelData
            screen: modelData
            visible: !GlobalStates.screenLocked

            readonly property bool reveal: !root.autoHide
                || islandHover.hovered || stripHover.hovered
                || dockApps.requestDockShow
                || !(ToplevelManager.activeToplevel?.activated ?? false)

            anchors {
                bottom: true
                left: true
                right: true
            }

            // Shown for good: windows stop above the island. Auto-hidden:
            // windows use the whole screen and the island floats over them.
            exclusiveZone: root.autoHide ? 0 : implicitHeight - (Appearance.sizes.hyprlandGapsOut) - (Appearance.sizes.elevationMargin - Appearance.sizes.hyprlandGapsOut)

            implicitWidth: dockBackground.implicitWidth
            WlrLayershell.namespace: "quickshell:dock"
            color: "transparent"

            implicitHeight: (Config.options?.dock.height ?? 70) + Appearance.sizes.elevationMargin + Appearance.sizes.hyprlandGapsOut + root.islandLift

            // Input: the island while it is shown, only the strip at the edge
            // while it is hidden; the rest is click-through.
            mask: Region {
                item: dockRoot.reveal ? dockBackground : revealArea
            }

            Item {
                id: revealArea
                anchors {
                    bottom: parent.bottom
                    horizontalCenter: parent.horizontalCenter
                }
                width: dockBackground.implicitWidth
                height: root.revealStrip
                // Passive: hover is seen even over the buttons on top.
                HoverHandler {
                    id: stripHover
                }
            }

            Item { // The island
                id: dockBackground
                anchors {
                    top: parent.top
                    bottom: parent.bottom
                    horizontalCenter: parent.horizontalCenter
                }
                implicitWidth: dockRow.implicitWidth + root.islandPadding * 2
                width: implicitWidth

                // Slides off the bottom edge while hidden.
                opacity: dockRoot.reveal ? 1 : 0
                transform: Translate {
                    y: dockRoot.reveal ? 0 : dockRoot.implicitHeight
                    Behavior on y {
                        animation: Appearance.animation.elementMoveFast.numberAnimation.createObject(this)
                    }
                }
                Behavior on opacity {
                    animation: Appearance.animation.elementMoveFast.numberAnimation.createObject(this)
                }

                HoverHandler {
                    id: islandHover
                }

                StyledRectangularShadow {
                    target: dockVisualBackground
                }
                Rectangle { // The real rectangle that is visible
                    id: dockVisualBackground
                    property real margin: Appearance.sizes.elevationMargin
                    anchors.fill: parent
                    anchors.topMargin: Appearance.sizes.elevationMargin
                    anchors.bottomMargin: Appearance.sizes.hyprlandGapsOut + root.islandLift
                    color: Qt.rgba(Appearance.colors.colLayer0.r, Appearance.colors.colLayer0.g, Appearance.colors.colLayer0.b, 0.78)
                    border.width: 1
                    border.color: Qt.rgba(Appearance.colors.colLayer0Border.r, Appearance.colors.colLayer0Border.g, Appearance.colors.colLayer0Border.b, 0.60)
                    radius: root.islandRadius
                }

                RowLayout {
                    id: dockRow
                    anchors.top: parent.top
                    anchors.bottom: parent.bottom
                    anchors.bottomMargin: root.islandLift
                    anchors.horizontalCenter: parent.horizontalCenter
                    spacing: 3
                    property real padding: 5

                    DockApps {
                        id: dockApps
                        buttonPadding: dockRow.padding
                    }
                }
            }
        }
    }
}
