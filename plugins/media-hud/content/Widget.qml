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
    readonly property string title: service && service.title !== "" ? service.title : "No Media"
    readonly property string artist: service ? service.artist : ""
    readonly property bool isPlaying: service && service.isPlaying

    readonly property bool showArtist: {
        const settings = pluginApi ? pluginApi.pluginSettings : null;
        return settings && settings.showArtist !== undefined ? settings.showArtist : true;
    }

    readonly property string displayText: {
        if (!service || service.title === "") return "MEDIA";
        if (root.showArtist && root.artist !== "") return root.title + " — " + root.artist;
        return root.title;
    }

    implicitWidth: row.implicitWidth
    implicitHeight: Math.max(row.implicitHeight, 18 * root.s)

    Row {
        id: row
        anchors.verticalCenter: parent.verticalCenter
        spacing: 6 * root.s

        // Music note glyph
        Text {
            anchors.verticalCenter: parent.verticalCenter
            text: root.isPlaying ? "♪" : "■"
            color: root.isPlaying ? Theme.accent : Theme.dim
            font.family: Theme.font
            font.pixelSize: 13 * root.s
        }

        // Track title
        Text {
            anchors.verticalCenter: parent.verticalCenter
            text: root.displayText
            color: root.isPlaying ? Theme.bright : Theme.dim
            font.family: Theme.font
            font.pixelSize: 12 * root.s
            elide: Text.ElideRight
            width: root.widthBudget > 0 ? Math.min(implicitWidth, root.widthBudget) : implicitWidth
        }
    }

    MouseArea {
        anchors.fill: parent
        onClicked: if (root.pluginApi) root.pluginApi.togglePanel()
    }
}
