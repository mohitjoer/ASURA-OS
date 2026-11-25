import QtQuick
import QtQuick.Layouts
import qs.modules.common
import qs.modules.common.widgets

// A popup's header: its icon and title on the left and, when there is one, the
// headline value (a battery's percentage) right-aligned, the way the value rows
// below it align theirs.
RowLayout {
    id: root
    required property var icon
    required property var label
    property string badge: ""
    property color badgeColor: Appearance.colors.colOnLayer0
    spacing: 8

    MaterialSymbol {
        Layout.alignment: Qt.AlignVCenter
        fill: 1
        text: root.icon
        iconSize: Appearance.font.pixelSize.larger
        color: Appearance.colors.colPrimary
    }

    StyledText {
        Layout.alignment: Qt.AlignVCenter
        text: root.label
        font {
            weight: Font.DemiBold
            pixelSize: Appearance.font.pixelSize.normal
        }
        color: Appearance.colors.colOnLayer0
    }

    Item {
        Layout.fillWidth: true
        implicitWidth: 16
    }

    StyledText {
        visible: root.badge !== ""
        Layout.alignment: Qt.AlignVCenter
        text: root.badge
        font {
            weight: Font.DemiBold
            pixelSize: Appearance.font.pixelSize.normal
        }
        color: root.badgeColor
    }
}
