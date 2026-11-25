pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import Quickshell.Io

Item {
    id: svc

    property var pluginApi
    readonly property var settings: pluginApi ? pluginApi.pluginSettings : null

    property int cpuPercent: 0
    property int memPercent: 0
    property int cpuTemp: 0
    property int gpuPercent: 0

    Process {
        id: statsProc
        command: ["sh", "-c", "a=$(awk 'NR==1{print $2+$3+$4+$5+$6, $5+$6}' /proc/stat); sleep 0.2; b=$(awk 'NR==1{print $2+$3+$4+$5+$6, $5+$6}' /proc/stat); set -- $a; t1=$1; i1=$2; set -- $b; t2=$1; i2=$2; dt=$((t2-t1)); di=$((i2-i1)); [ \"$dt\" -gt 0 ] && cpu=$((100*(dt-di)/dt)) || cpu=0; mem=$(awk '/^MemTotal:/{t=$2} /^MemAvailable:/{a=$2} END{printf \"%d\", (t-a)*100/t}' /proc/meminfo); temp=$(cat /sys/class/thermal/thermal_zone0/temp 2>/dev/null || echo 0); temp=$((temp/1000)); printf '{\"cpu\":%d,\"mem\":%d,\"temp\":%d}\\n' \"$cpu\" \"$mem\" \"$temp\""]
        stdout: StdioCollector { id: statsOut }
        onExited: (code) => {
            if (code === 0 && statsOut.text.length > 0) {
                try {
                    const data = JSON.parse(statsOut.text);
                    svc.cpuPercent = Math.round(data.cpu || 0);
                    svc.memPercent = Math.round(data.mem || 0);
                    svc.cpuTemp = Math.round(data.temp || 0);
                } catch (e) {
                }
            }
        }
    }

    Timer {
        interval: 2000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: {
            if (!statsProc.running) {
                statsProc.running = true;
            }
        }
    }
}
