pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.PluginKit.Singletons

Item {
    id: root

    property var pluginApi
    property string density: "full"
    property real s: 1
    property real widthBudget: 320
    property bool active: false

    readonly property var service: pluginApi ? pluginApi.mainInstance : null
    readonly property int cpu: service ? service.cpuPercent : 0
    readonly property int mem: service ? service.memPercent : 0
    readonly property int temp: service ? service.cpuTemp : 0
    readonly property int gpu: service ? service.gpuPercent : 0

    implicitWidth: root.widthBudget
    implicitHeight: col.implicitHeight + 24 * root.s

    Column {
        id: col
        x: 12 * root.s
        y: 12 * root.s
        width: root.width - 24 * root.s
        spacing: 12 * root.s

        // Header
        Row {
            width: parent.width
            Text {
                text: "SYSTEM VITALS"
                color: Theme.bright
                font.family: Theme.display
                font.pixelSize: 15 * root.s
                font.weight: Font.Bold
            }
        }

        // CPU Metric
        Column {
            width: parent.width
            spacing: 4 * root.s
            Row {
                width: parent.width
                Text { text: "CPU LOAD"; color: Theme.dim; font.family: Theme.font; font.pixelSize: 11 * root.s }
                Item { width: 1; height: 1 }
                Text { text: root.cpu + "%" + (root.temp > 0 ? (" (" + root.temp + "°C)") : ""); color: Theme.bright; font.family: Theme.font; font.pixelSize: 11 * root.s; x: col.width - implicitWidth }
            }
            Rectangle {
                width: parent.width
                height: 6 * root.s
                radius: 3 * root.s
                color: Theme.hair
                Rectangle {
                    width: Math.min(parent.width, parent.width * (root.cpu / 100.0))
                    height: parent.height
                    radius: 3 * root.s
                    color: Theme.bright
                }
            }
        }

        // Memory Metric
        Column {
            width: parent.width
            spacing: 4 * root.s
            Row {
                width: parent.width
                Text { text: "MEMORY (RAM)"; color: Theme.dim; font.family: Theme.font; font.pixelSize: 11 * root.s }
                Text { text: root.mem + "%"; color: Theme.bright; font.family: Theme.font; font.pixelSize: 11 * root.s; x: col.width - implicitWidth }
            }
            Rectangle {
                width: parent.width
                height: 6 * root.s
                radius: 3 * root.s
                color: Theme.hair
                Rectangle {
                    width: Math.min(parent.width, parent.width * (root.mem / 100.0))
                    height: parent.height
                    radius: 3 * root.s
                    color: Theme.bright
                }
            }
        }

        // GPU Metric (if available)
        Column {
            width: parent.width
            spacing: 4 * root.s
            Row {
                width: parent.width
                Text { text: "GPU UTILIZATION"; color: Theme.dim; font.family: Theme.font; font.pixelSize: 11 * root.s }
                Text { text: root.gpu + "%"; color: Theme.bright; font.family: Theme.font; font.pixelSize: 11 * root.s; x: col.width - implicitWidth }
            }
            Rectangle {
                width: parent.width
                height: 6 * root.s
                radius: 3 * root.s
                color: Theme.hair
                Rectangle {
                    width: Math.min(parent.width, parent.width * (root.gpu / 100.0))
                    height: parent.height
                    radius: 3 * root.s
                    color: Theme.bright
                }
            }
        }

        // Launch full sysmon action button
        Rectangle {
            width: parent.width
            height: 28 * root.s
            radius: Theme.radius
            color: openArea.pressed ? Theme.hair : "transparent"
            border.width: 1
            border.color: Theme.lineStrong

            Text {
                anchors.centerIn: parent
                text: "OPEN RESOURCE MONITOR"
                color: Theme.bright
                font.family: Theme.font
                font.pixelSize: 11 * root.s
                font.weight: Font.Medium
            }

            MouseArea {
                id: openArea
                anchors.fill: parent
                onClicked: {
                    Quickshell.execDetached(["asura-sysmon", "ui"]);
                    if (root.pluginApi)
                        root.pluginApi.closePanel();
                }
            }
        }
    }
}
