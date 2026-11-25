import QtQuick
import qs.PluginKit.Singletons

Item {
    id: root

    property var pluginApi
    property var screen
    property bool active: false
    property string density: "glyph"
    property real s: 1
    property real widthBudget: 0

    readonly property var service: pluginApi ? pluginApi.mainInstance : null
    readonly property int cpu: service ? service.cpuPercent : 0
    readonly property int mem: service ? service.memPercent : 0
    readonly property int temp: service ? service.cpuTemp : 0

    readonly property bool showTemp: {
        const settings = pluginApi ? pluginApi.pluginSettings : null;
        return settings && settings.showTemp !== undefined ? settings.showTemp : true;
    }

    implicitWidth: row.implicitWidth
    implicitHeight: Math.max(row.implicitHeight, 18 * root.s)

    Row {
        id: row
        anchors.verticalCenter: parent.verticalCenter
        spacing: 6 * root.s

        // Pulse Mark
        Item {
            anchors.verticalCenter: parent.verticalCenter
            width: 12 * root.s
            height: 12 * root.s

            Rectangle {
                anchors.centerIn: parent
                width: 7 * root.s
                height: 7 * root.s
                rotation: 45
                color: root.active ? Theme.accent : Theme.bright
            }
        }

        // Metrics readout
        Text {
            anchors.verticalCenter: parent.verticalCenter
            text: "CPU " + root.cpu + "%" + "  RAM " + root.mem + "%" + (root.showTemp && root.temp > 0 ? ("  " + root.temp + "°C") : "")
            color: Theme.bright
            font.family: Theme.font
            font.pixelSize: 12 * root.s
            font.weight: Font.Medium
            elide: Text.ElideRight
            width: root.widthBudget > 0 ? Math.min(implicitWidth, root.widthBudget) : implicitWidth
        }
    }

    MouseArea {
        anchors.fill: parent
        onClicked: if (root.pluginApi) root.pluginApi.togglePanel()
    }
}
