pragma ComponentBehavior: Bound

// The battery widget for the desktop.
// Shows battery percentage and state (charging/discharging)
// Read from /sys/class/power_supply or /proc/acpi/battery.

import QtQuick
import Quickshell.Io
import Quickshell.Widgets
import "../../../components"
import "../Singletons"

Item {
    id: root

    property string style: "default"
    property real s: 1.0
    property bool active: true
    property int refreshSeconds: 5

    // Read from sysfs power supply
    FileView {
        id: capacityFile
        path: "/sys/class/power_supply/BAT0/capacity"
        watchChanges: false
    }

    FileView {
        id: statusFile
        path: "/sys/class/power_supply/BAT0/status"
        watchChanges: false
    }

    FileView {
        id: presentFile
        path: "/sys/class/power_supply/BAT0/present"
        watchChanges: false
    }

    Timer {
        interval: root.refreshSeconds * 1000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: {
            capacityFile.reload();
            statusFile.reload();
            presentFile.reload();
        }
    }

    readonly property int percent: {
        var cap = capacityFile.text().trim();
        if (/^\d+$/.test(cap)) {
            return parseInt(cap, 10);
        }
        return 0;
    }

    readonly property string state: {
        var status = statusFile.text().trim().toLowerCase();
        if (status === "charging" || status === "full") {
            return status;
        }
        return "discharging";
    }

    readonly property bool isCharging: state === "charging" || state === "full"

    Text {
        id: label
        anchors.fill: parent
        text: percent + "%"
        color: isCharging ? Theme.inkPrimary : (Config.dark ? Theme.inkPrimary : Theme.inkSecondary)
        font.pixelSize: 13
        font.weight: Font.DemiBold
        horizontalAlignment: Text.AlignHCenter
        verticalAlignment: Text.AlignVCenter
    }

    GlyphIcon {
        anchors { verticalCenter: label.verticalCenter; horizontalCenter: label.horizontalCenter }
        width: 16
        height: 16
        color: isCharging ? Theme.inkPrimary : (Config.dark ? Theme.inkPrimary : Theme.inkSecondary)
        visible: isCharging
        name: "bolt"
    }

    GlyphIcon {
        anchors { verticalCenter: label.verticalCenter; horizontalCenter: label.horizontalCenter }
        width: 16
        height: 16
        color: !isCharging && percent < 20 ? Theme.colorError : (Config.dark ? Theme.inkPrimary : Theme.inkSecondary)
        visible: !isCharging
        name: "battery"
    }
}
