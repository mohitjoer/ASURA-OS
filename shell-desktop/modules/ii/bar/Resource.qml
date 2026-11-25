import qs.modules.common
import qs.modules.common.widgets
import QtQuick
import QtQuick.Layouts

// ASURA overlay (shell-desktop/asura/vendor.sh OVERLAYS): one resource readout,
// in the professional style of the ASURA bar: a short label ("CPU"), the
// percentage in tabular figures, and a thin usage bar under them that turns
// the error colour past the warning threshold. Same properties as end-4's
// (iconName, percentage, warningThreshold, shown), so the vertical bar keeps
// working; `label` names the readout, falling back to iconName.
Item {
    id: root
    required property string iconName
    required property double percentage
    property string label: ""
    property int warningThreshold: 100
    property bool shown: true
    readonly property bool warning: percentage * 100 >= warningThreshold
    readonly property color accent: warning ? Appearance.colors.colError : Appearance.colors.colPrimary

    visible: shown
    implicitWidth: shown ? content.implicitWidth : 0
    implicitHeight: Appearance.sizes.barHeight

    ColumnLayout {
        id: content
        anchors.centerIn: parent
        spacing: 2

        RowLayout {
            spacing: 5
            StyledText {
                text: root.label !== "" ? root.label : root.iconName
                color: Appearance.colors.colOnLayer1Inactive
                font.pixelSize: Appearance.font.pixelSize.smallest
                font.weight: Font.DemiBold
                font.letterSpacing: 0.8
                font.capitalization: Font.AllUppercase
            }
            StyledText {
                // Fixed width and a monospace font, so the bar does not shift as numbers change.
                Layout.preferredWidth: widest.width
                horizontalAlignment: Text.AlignRight
                text: `${Math.round(root.percentage * 100)}%`
                color: root.warning ? Appearance.colors.colError : Appearance.colors.colOnLayer1
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

        Rectangle { // usage bar
            Layout.fillWidth: true
            implicitHeight: 2
            color: Appearance.colors.colLayer1Hover
            Rectangle {
                width: parent.width * Math.max(0, Math.min(1, root.percentage))
                height: parent.height
                color: root.accent
                Behavior on width {
                    animation: Appearance.animation.elementMoveFast.numberAnimation.createObject(this)
                }
            }
        }
    }
}
