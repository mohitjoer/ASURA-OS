pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Controls
import Quickshell
import Quickshell.Io
import qs.Ui
import qs.Ui.Singletons
import "../Singletons"

// About Page.
// Distribution overview, OS edition, hardware specifications, and project links.
Item {
    id: pg

    property var hub
    readonly property bool fullBleed: true

    property string hardwareModel: "Standard PC"
    Process {
        id: hwModelProc
        running: true
        command: ["sh", "-c", "cat /sys/class/dmi/id/product_name 2>/dev/null || cat /sys/devices/virtual/dmi/id/product_name 2>/dev/null || cat /etc/hostname 2>/dev/null || echo 'Standard PC'"]
        stdout: StdioCollector {
            onStreamFinished: {
                var s = text.trim();
                if (s.length > 0)
                    pg.hardwareModel = s;
            }
        }
    }

    // ── Head ────────────────────────────────────────────────────────────────
    Column {
        id: head
        anchors {
            left: parent.left; right: parent.right; top: parent.top
            leftMargin: Tokens.s6; rightMargin: Tokens.s6; topMargin: Tokens.s6
        }
        spacing: Tokens.s3

        Row {
            height: Tokens.s5
            spacing: Tokens.s2
            Rectangle {
                width: 16; height: 1; color: Tokens.ink
                anchors.verticalCenter: parent.verticalCenter
            }
            Image {
                width: 12; height: 12
                anchors.verticalCenter: parent.verticalCenter
                source: Qt.resolvedUrl(Tokens.light ? "../../assets/toren-logo.svg" : "../../assets/toren-logo-on-dark.svg")
                sourceSize.width: 24; sourceSize.height: 24
                fillMode: Image.PreserveAspectFit
                smooth: true
            }
            Text {
                text: I18n.tr("ABOUT"); color: Tokens.inkMuted; font.family: Tokens.ui
                font.pixelSize: 9; font.weight: Font.Medium; font.letterSpacing: Tokens.trackMark
                anchors.verticalCenter: parent.verticalCenter
            }
        }

        Text {
            text: I18n.tr("About Toren Linux"); color: Tokens.ink
            font.family: Tokens.display; font.pixelSize: Tokens.fTitle
        }

        Text {
            width: Math.min(parent.width, 720)
            text: I18n.tr("Pure monochrome desktop OS, engineered for precision, speed, and minimalism.")
            color: Tokens.inkMuted; font.family: Tokens.ui
            font.pixelSize: Tokens.fBody; wrapMode: Text.WordWrap
        }
    }

    // ── Content ─────────────────────────────────────────────────────────────
    Flickable {
        id: flick
        anchors {
            left: parent.left; right: parent.right
            top: head.bottom; bottom: parent.bottom
            leftMargin: Tokens.s6; rightMargin: Tokens.s6
            topMargin: Tokens.s5; bottomMargin: Tokens.s6
        }
        contentWidth: width
        contentHeight: Math.max(col.height + Tokens.s6, height)
        clip: true
        boundsBehavior: Flickable.StopAtBounds
        ScrollBar.vertical: ScrollRail { policy: ScrollBar.AsNeeded }
        WheelScroll { }

        CardColumns {
            id: col
            width: flick.width - Tokens.s3
            spacing: Tokens.s5
            fillTo: flick.height
            minColumnWidth: 320
            maxColumns: 2

            // ── Card 1: System & Distribution ──
            SettingCard {
                width: col.colWidth
                title: I18n.tr("SYSTEM & EDITION")
                collapsible: false

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: false
                    label: I18n.tr("Distribution")
                    value: "Toren Linux"
                    desc: I18n.tr("Independent monochrome Wayland desktop OS")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Version")
                    value: Updates.currentVersion.length > 0 ? Updates.currentVersion : "Rolling (dev-unstable)"
                    desc: I18n.tr("Active desktop release channel")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Base System")
                    value: "Void Linux"
                    desc: I18n.tr("Continuous release base distribution (xbps)")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Desktop & Compositor")
                    value: "Quickshell \u00b7 Hyprland"
                    desc: I18n.tr("Custom reactive shell on Wayland")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Kernel")
                    value: SysInfo.sysKernel.length > 0 && SysInfo.sysKernel !== "-" ? SysInfo.sysKernel : "Linux"
                    desc: SysInfo.sysArch.length > 0 ? SysInfo.sysArch : "x86_64"
                }
            }

            // ── Card 2: Hardware Platform ──
            SettingCard {
                width: col.colWidth
                title: I18n.tr("HARDWARE PLATFORM")
                collapsible: false

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: false
                    label: I18n.tr("Model")
                    value: pg.hardwareModel
                    desc: I18n.tr("Device hardware platform")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Processor")
                    value: SysInfo.sysCpuCores.length > 0 && SysInfo.sysCpuCores !== "-" ? (SysInfo.sysCpuCores + " cores") : ""
                    desc: SysInfo.sysCpu.length > 0 && SysInfo.sysCpu !== "-" ? SysInfo.sysCpu : I18n.tr("x86_64 Processor")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Memory")
                    value: SysInfo.sysRamPct + "% used"
                    desc: SysInfo.sysRam.length > 0 && SysInfo.sysRam !== "- / -" ? SysInfo.sysRam : I18n.tr("System RAM")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Graphics")
                    value: SysInfo.sysRefresh.length > 0 ? SysInfo.sysRefresh : ""
                    desc: SysInfo.sysGpu.length > 0 && SysInfo.sysGpu !== "-"
                        ? SysInfo.sysGpu + (SysInfo.sysGpu2.length > 0 ? " / " + SysInfo.sysGpu2 : "")
                        : (SysInfo.sysResolution.length > 0 && SysInfo.sysResolution !== "-" ? SysInfo.sysResolution : I18n.tr("Integrated Graphics"))
                }
            }

            // ── Card 3: Project & Links ──
            SettingCard {
                width: col.colWidth
                title: I18n.tr("PROJECT & REPOSITORY")
                collapsible: false

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: false
                    label: I18n.tr("Source Code")
                    value: "GitHub"
                    desc: "github.com/mohitjoer/Toren-Linux"
                    MouseArea {
                        anchors.fill: parent
                        cursorShape: Qt.PointingHandCursor
                        onClicked: Qt.openUrlExternally("https://github.com/mohitjoer/Toren-Linux")
                    }
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Issue Tracker")
                    value: I18n.tr("Bug Reports")
                    desc: "github.com/mohitjoer/Toren-Linux/issues"
                    MouseArea {
                        anchors.fill: parent
                        cursorShape: Qt.PointingHandCursor
                        onClicked: Qt.openUrlExternally("https://github.com/mohitjoer/Toren-Linux/issues")
                    }
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Documentation")
                    value: I18n.tr("User Manual")
                    desc: "github.com/mohitjoer/Toren-Linux/tree/main/docs"
                    MouseArea {
                        anchors.fill: parent
                        cursorShape: Qt.PointingHandCursor
                        onClicked: Qt.openUrlExternally("https://github.com/mohitjoer/Toren-Linux/tree/main/docs")
                    }
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("License")
                    value: "GPL-3.0"
                    desc: I18n.tr("GNU General Public License v3.0 or later")
                }
            }

            // ── Card 4: Acknowledgements ──
            SettingCard {
                width: col.colWidth
                title: I18n.tr("ACKNOWLEDGEMENTS")
                collapsible: false

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: false
                    label: I18n.tr("Open Source Foundations")
                    desc: I18n.tr("Toren is built on the shoulders of giants: Void Linux, Quickshell, Hyprland, Wayland, and the open source community.")
                }

                SettingRow {
                    anchors.left: parent.left; anchors.right: parent.right
                    divider: true
                    label: I18n.tr("Contributors & Testers")
                    desc: I18n.tr("Gratitude to all contributors, developers, and testers crafting, polishing, and testing Toren.")
                }
            }
        }
    }
}
