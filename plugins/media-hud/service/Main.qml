pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import Quickshell.Io

Item {
    id: svc

    property var pluginApi
    readonly property var settings: pluginApi ? pluginApi.pluginSettings : null

    property string title: ""
    property string artist: ""
    property string status: "Stopped"
    property bool isPlaying: status === "Playing"

    Process {
        id: statusProc
        command: ["playerctl", "status"]
        stdout: SplitParser {
            onRead: (line) => {
                svc.status = line.trim();
            }
        }
    }

    Process {
        id: titleProc
        command: ["playerctl", "metadata", "title"]
        stdout: SplitParser {
            onRead: (line) => {
                svc.title = line.trim();
            }
        }
    }

    Process {
        id: artistProc
        command: ["playerctl", "metadata", "artist"]
        stdout: SplitParser {
            onRead: (line) => {
                svc.artist = line.trim();
            }
        }
    }

    function togglePlay() {
        Quickshell.execDetached(["playerctl", "play-pause"]);
        refresh();
    }

    function nextTrack() {
        Quickshell.execDetached(["playerctl", "next"]);
        refresh();
    }

    function prevTrack() {
        Quickshell.execDetached(["playerctl", "previous"]);
        refresh();
    }

    function refresh() {
        if (!statusProc.running) statusProc.running = true;
        if (!titleProc.running) titleProc.running = true;
        if (!artistProc.running) artistProc.running = true;
    }

    Timer {
        interval: 2500
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: svc.refresh()
    }
}
