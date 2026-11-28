pragma ComponentBehavior: Bound

import qs.services
import qs.modules.common
import qs.modules.common.functions
import qs.modules.common.widgets
import QtQuick
import QtQuick.Layouts
import Quickshell

Item {
    id: root
    property real padding: 4
    implicitWidth: Math.min(Math.max(flow.implicitWidth + 16, 720), (QsWindow?.window?.screen?.width ?? 1280) * 0.88)
    implicitHeight: (QsWindow?.window?.screen?.height ?? 800) * 0.80

    StyledFlickable {
        id: flickable
        clip: true
        anchors.fill: parent
        anchors.margins: Appearance.rounding.small
        contentHeight: Math.max(height, flow.childrenRect.height + 16)
        contentWidth: Math.max(width, flow.implicitWidth)
        Flow {
            id: flow
            height: flickable.height
            flow: Flow.TopToBottom
            spacing: 12
            Repeater {
                model: [...HyprlandKeybinds.keybindCategories, ""]
                delegate: CheatsheetKeybindsCategory {
                    required property var modelData
                    categoryName: modelData
                }
            }
        }
    }

    ScrollEdgeFade {
        target: flickable
        vertical: false
        color: Appearance.colors.colLayer0Base
    }
}
