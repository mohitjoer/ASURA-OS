import QtQuick
import QtQuick.Layouts
import qs.modules.common
import qs.modules.common.widgets

RowLayout {
    id: root
    required property string icon
    required property string label
    required property string value
    property color valueColor: Appearance.colors.colOnLayer0
    spacing: 8

    MaterialSymbol {
        text: root.icon
        color: Appearance.colors.colSubtext
        iconSize: Appearance.font.pixelSize.medium
        Layout.alignment: Qt.AlignVCenter
    }
    StyledText {
        text: root.label
        color: Appearance.colors.colSubtext
        font.pixelSize: Appearance.font.pixelSize.small
        Layout.alignment: Qt.AlignVCenter
    }
    Item {
        Layout.fillWidth: true
        implicitWidth: 16
    }
    StyledText {
        Layout.alignment: Qt.AlignVCenter
        horizontalAlignment: Text.AlignRight
        visible: root.value !== ""
        font {
            weight: Font.DemiBold
            pixelSize: Appearance.font.pixelSize.small
        }
        color: root.valueColor
        text: root.value
    }
}
