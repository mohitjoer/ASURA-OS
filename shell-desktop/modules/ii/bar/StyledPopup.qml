import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.functions
import QtQuick
import QtQuick.Effects
import Quickshell
import Quickshell.Wayland

LazyLoader {
    id: root

    property Item hoverTarget
    default property Item contentItem
    property real popupBackgroundMargin: 0

    active: !!(hoverTarget && hoverTarget.containsMouse)

    component: PanelWindow {
        id: popupWindow
        color: "transparent"
        screen: root.hoverTarget?.QsWindow?.window?.screen ?? root.QsWindow?.window?.screen ?? null

        anchors.left: !Config.options.bar.vertical || (Config.options.bar.vertical && !Config.options.bar.bottom)
        anchors.right: Config.options.bar.vertical && Config.options.bar.bottom
        anchors.top: Config.options.bar.vertical || (!Config.options.bar.vertical && !Config.options.bar.bottom)
        anchors.bottom: !Config.options.bar.vertical && Config.options.bar.bottom

        implicitWidth: popupBackground.implicitWidth + Appearance.sizes.elevationMargin * 2 + root.popupBackgroundMargin
        implicitHeight: popupBackground.implicitHeight + Appearance.sizes.elevationMargin * 2 + root.popupBackgroundMargin

        mask: Region {
            item: popupBackground
        }

        exclusionMode: ExclusionMode.Ignore
        exclusiveZone: 0
        margins {
            left: {
                if (!Config.options.bar.vertical) {
                    let win = root.hoverTarget?.QsWindow ?? root.QsWindow;
                    let targetScreen = win?.window?.screen ?? popupWindow.screen;
                    let screenWidth = targetScreen?.width ?? 1920;
                    let targetX = (root.hoverTarget && win) ? win.mapFromItem(
                        root.hoverTarget, 
                        root.hoverTarget.width / 2, 0
                    ).x : 0;
                    let idealLeft = targetX - (popupWindow.implicitWidth / 2);
                    let edgeMargin = 12;
                    let minLeft = edgeMargin;
                    let maxLeft = screenWidth - popupWindow.implicitWidth - edgeMargin;
                    if (maxLeft < minLeft) return minLeft;
                    return Math.max(minLeft, Math.min(idealLeft, maxLeft));
                }
                return Appearance.sizes.verticalBarWidth + 4;
            }
            top: {
                if (!Config.options.bar.vertical) {
                    return Appearance.sizes.barHeight + 4;
                }
                let win = root.hoverTarget?.QsWindow ?? root.QsWindow;
                let targetScreen = win?.window?.screen ?? popupWindow.screen;
                let screenHeight = targetScreen?.height ?? 1080;
                let targetY = (root.hoverTarget && win) ? win.mapFromItem(
                    root.hoverTarget, 
                    0, root.hoverTarget.height / 2
                ).y : 0;
                let idealTop = targetY - (popupWindow.implicitHeight / 2);
                let edgeMargin = 12;
                let minTop = edgeMargin;
                let maxTop = screenHeight - popupWindow.implicitHeight - edgeMargin;
                if (maxTop < minTop) return minTop;
                return Math.max(minTop, Math.min(idealTop, maxTop));
            }
            right: Appearance.sizes.verticalBarWidth + 4
            bottom: Appearance.sizes.barHeight + 4
        }
        WlrLayershell.namespace: "quickshell:popup"
        WlrLayershell.layer: WlrLayer.Overlay

        StyledRectangularShadow {
            target: popupBackground
        }

        Rectangle {
            id: popupBackground
            readonly property real margin: 12
            anchors {
                fill: parent
                leftMargin: Appearance.sizes.elevationMargin + root.popupBackgroundMargin * (!popupWindow.anchors.left)
                rightMargin: Appearance.sizes.elevationMargin + root.popupBackgroundMargin * (!popupWindow.anchors.right)
                topMargin: Appearance.sizes.elevationMargin + root.popupBackgroundMargin * (!popupWindow.anchors.top)
                bottomMargin: Appearance.sizes.elevationMargin + root.popupBackgroundMargin * (!popupWindow.anchors.bottom)
            }
            implicitWidth: root.contentItem.implicitWidth + margin * 2
            implicitHeight: root.contentItem.implicitHeight + margin * 2
            color: Appearance.colors.colLayer0
            radius: Appearance.rounding.windowRounding
            children: [root.contentItem]

            border.width: 1
            border.color: Appearance.colors.colLayer0Border
        }
    }
}
