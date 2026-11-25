import qs.modules.common
import qs.modules.common.widgets
import qs.services
import QtQuick
import QtQuick.Layouts

// ASURA overlay (shell-desktop/toren/vendor.sh OVERLAYS): the battery, at the very
// end of the bar, in the style of the resources: a drawn battery whose fill is
// the charge, and the percentage in a fixed-width monospace figure. Charging
// shows a bolt and the accent colour; low and not charging turns the error
// colour. Hovering shows end-4's battery popup.
MouseArea {
    id: root
    readonly property bool isCharging: Battery.isCharging
    readonly property bool isPluggedIn: Battery.isPluggedIn
    readonly property real percentage: Battery.percentage
    readonly property bool isLow: percentage <= Config.options.battery.low / 100
    readonly property color levelColor: (isLow && !isCharging) ? Appearance.colors.colError
        : isCharging ? Appearance.colors.colPrimary
        : Appearance.colors.colOnLayer1

    implicitWidth: row.implicitWidth + 4
    implicitHeight: Appearance.sizes.barHeight
    hoverEnabled: !Config.options.bar.tooltips.clickToShow

    RowLayout {
        id: row
        anchors.centerIn: parent
        spacing: 6

        Item { // the battery
            implicitWidth: 24
            implicitHeight: 12

            Rectangle { // body
                id: body
                width: 22
                height: 12
                color: "transparent"
                border.width: 1.5
                border.color: Appearance.colors.colOnLayer1Inactive
                radius: 2

                Rectangle { // charge
                    x: 2.5
                    y: 2.5
                    height: parent.height - 5
                    width: Math.max(1, (parent.width - 5) * Math.max(0, Math.min(1, root.percentage)))
                    color: root.levelColor
                    radius: 1
                    Behavior on width {
                        animation: Appearance.animation.elementMoveFast.numberAnimation.createObject(this)
                    }
                }
            }
            Rectangle { // terminal
                anchors {
                    left: body.right
                    verticalCenter: body.verticalCenter
                }
                width: 2
                height: 5
                color: Appearance.colors.colOnLayer1Inactive
            }
            MaterialSymbol { // charging
                anchors.centerIn: body
                visible: root.isCharging
                fill: 1
                text: "bolt"
                iconSize: 13
                color: Appearance.colors.colOnLayer0
                style: Text.Outline
                styleColor: Appearance.colors.colLayer0
            }
        }

        StyledText {
            Layout.preferredWidth: widest.width
            horizontalAlignment: Text.AlignRight
            text: `${Math.round(root.percentage * 100)}%`
            color: (root.isLow && !root.isCharging) ? Appearance.colors.colError : Appearance.colors.colOnLayer1
            font.pixelSize: Appearance.font.pixelSize.small
            font.family: Appearance.font.family.monospace
            TextMetrics {
                id: widest
                text: "100%"
                font.pixelSize: Appearance.font.pixelSize.small
                font.family: Appearance.font.family.monospace
            }
        }
    }

    BatteryPopup {
        id: batteryPopup
        hoverTarget: root
    }
}
