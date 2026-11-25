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
    readonly property string title: service && service.title !== "" ? service.title : "Not Playing"
    readonly property string artist: service && service.artist !== "" ? service.artist : "Unknown Artist"
    readonly property bool isPlaying: service && service.isPlaying

    implicitWidth: root.widthBudget
    implicitHeight: col.implicitHeight + 24 * root.s

    Column {
        id: col
        x: 12 * root.s
        y: 12 * root.s
        width: root.width - 24 * root.s
        spacing: 12 * root.s

        Text {
            text: "NOW PLAYING"
            color: Theme.dim
            font.family: Theme.display
            font.pixelSize: 11 * root.s
            font.letterSpacing: 1.5
        }

        // Track & Artist
        Column {
            width: parent.width
            spacing: 3 * root.s

            Text {
                width: parent.width
                text: root.title
                color: Theme.bright
                font.family: Theme.display
                font.pixelSize: 16 * root.s
                font.weight: Font.Bold
                elide: Text.ElideRight
            }

            Text {
                width: parent.width
                text: root.artist
                color: Theme.dim
                font.family: Theme.font
                font.pixelSize: 12 * root.s
                elide: Text.ElideRight
            }
        }

        // Playback Controls Row
        Row {
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: 16 * root.s

            // Prev
            Rectangle {
                width: 32 * root.s
                height: 32 * root.s
                radius: Theme.radius
                color: prevArea.pressed ? Theme.hair : "transparent"
                border.width: 1
                border.color: Theme.lineStrong

                Text {
                    anchors.centerIn: parent
                    text: "⏮"
                    color: Theme.bright
                    font.family: Theme.font
                    font.pixelSize: 13 * root.s
                }

                MouseArea {
                    id: prevArea
                    anchors.fill: parent
                    onClicked: if (root.service) root.service.prevTrack()
                }
            }

            // Play / Pause
            Rectangle {
                width: 44 * root.s
                height: 32 * root.s
                radius: Theme.radius
                color: playArea.pressed ? Theme.hair : Theme.hair
                border.width: 1
                border.color: Theme.bright

                Text {
                    anchors.centerIn: parent
                    text: root.isPlaying ? "⏸" : "▶"
                    color: Theme.bright
                    font.family: Theme.font
                    font.pixelSize: 14 * root.s
                }

                MouseArea {
                    id: playArea
                    anchors.fill: parent
                    onClicked: if (root.service) root.service.togglePlay()
                }
            }

            // Next
            Rectangle {
                width: 32 * root.s
                height: 32 * root.s
                radius: Theme.radius
                color: nextArea.pressed ? Theme.hair : "transparent"
                border.width: 1
                border.color: Theme.lineStrong

                Text {
                    anchors.centerIn: parent
                    text: "⏭"
                    color: Theme.bright
                    font.family: Theme.font
                    font.pixelSize: 13 * root.s
                }

                MouseArea {
                    id: nextArea
                    anchors.fill: parent
                    onClicked: if (root.service) root.service.nextTrack()
                }
            }
        }
    }
}
