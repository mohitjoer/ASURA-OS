import qs.services
import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.functions
import QtQuick
import QtQuick.Layouts
import Quickshell

RippleButton {
    id: lightDarkButtonRoot
    required property bool dark
    padding: 5
    Layout.fillWidth: true
    colBackground: Appearance.colors.colLayer2
    toggled: Appearance.m3colors.darkmode === dark
    onClicked: {
        Quickshell.execDetached(["bash", "-c", `${Directories.wallpaperSwitchScriptPath} --mode ${dark ? "dark" : "light"} --noswitch`])
    }
    contentItem: Item {
        anchors.centerIn: parent
        implicitWidth: buttonContentLayout.implicitWidth
        implicitHeight: buttonContentLayout.implicitHeight
        ColumnLayout {
            id: buttonContentLayout
            anchors.centerIn: parent
            spacing: 8

            // ASURA OS Desktop Preview
            Rectangle {
                Layout.alignment: Qt.AlignHCenter
                implicitWidth: 260
                implicitHeight: 145
                radius: lightDarkButtonRoot.buttonRadius - lightDarkButtonRoot.padding
                color: lightDarkButtonRoot.dark ? "#09090c" : "#f5f4f0"
                clip: true
                border {
                    width: lightDarkButtonRoot.toggled ? 2 : 1
                    color: lightDarkButtonRoot.toggled ? Appearance.m3colors.m3primary : (lightDarkButtonRoot.dark ? "#26262a" : "#d8d6ce")
                }

                // Mini Asura Top Bar
                Rectangle {
                    id: miniBar
                    anchors.top: parent.top
                    anchors.left: parent.left
                    anchors.right: parent.right
                    height: 22
                    color: lightDarkButtonRoot.dark ? "#121216" : "#e8e6e1"
                    border.width: 1
                    border.color: lightDarkButtonRoot.dark ? "#1e1e24" : "#dedcd6"

                    Row {
                        anchors.left: parent.left
                        anchors.leftMargin: 8
                        anchors.verticalCenter: parent.verticalCenter
                        spacing: 5

                        // Signature Asura Mark //
                        Text {
                            text: "//"
                            color: "#e2342a"
                            font.family: Appearance.font.family.title
                            font.pixelSize: 11
                            font.weight: Font.Black
                            anchors.verticalCenter: parent.verticalCenter
                        }

                        // Workspaces
                        Rectangle {
                            width: 14
                            height: 14
                            radius: 3
                            color: "#e2342a"
                            anchors.verticalCenter: parent.verticalCenter
                            Text {
                                anchors.centerIn: parent
                                text: "1"
                                color: "#ffffff"
                                font.pixelSize: 9
                                font.weight: Font.Bold
                            }
                        }
                        Text {
                            text: "2"
                            color: lightDarkButtonRoot.dark ? "#666670" : "#999990"
                            font.pixelSize: 9
                            anchors.verticalCenter: parent.verticalCenter
                        }
                        Text {
                            text: "3"
                            color: lightDarkButtonRoot.dark ? "#666670" : "#999990"
                            font.pixelSize: 9
                            anchors.verticalCenter: parent.verticalCenter
                        }
                    }

                    // Center: TOREN
                    Text {
                        anchors.centerIn: parent
                        text: "TOREN"
                        color: lightDarkButtonRoot.dark ? "#888894" : "#777770"
                        font.family: Appearance.font.family.main
                        font.pixelSize: 9
                        font.weight: Font.DemiBold
                    }

                    // Right: System Tray Glyphs
                    Row {
                        anchors.right: parent.right
                        anchors.rightMargin: 8
                        anchors.verticalCenter: parent.verticalCenter
                        spacing: 4

                        Rectangle {
                            width: 6; height: 6; radius: 3
                            color: "#2ecc71"
                            anchors.verticalCenter: parent.verticalCenter
                        }
                        Text {
                            text: "12:00"
                            color: lightDarkButtonRoot.dark ? "#aaaabb" : "#555550"
                            font.pixelSize: 8
                            anchors.verticalCenter: parent.verticalCenter
                        }
                    }
                }

                // Mini Desktop Area with Floating Asura Window
                Item {
                    anchors.top: miniBar.bottom
                    anchors.left: parent.left
                    anchors.right: parent.right
                    anchors.bottom: parent.bottom

                    Rectangle {
                        anchors.centerIn: parent
                        width: parent.width - 24
                        height: parent.height - 18
                        radius: 4
                        color: lightDarkButtonRoot.dark ? "#040406" : "#ffffff"
                        border.width: 1
                        border.color: lightDarkButtonRoot.dark ? "#22222a" : "#dcd9d2"

                        // Window Titlebar
                        Rectangle {
                            id: winTitlebar
                            anchors.top: parent.top
                            anchors.left: parent.left
                            anchors.right: parent.right
                            height: 16
                            radius: 3
                            color: lightDarkButtonRoot.dark ? "#0e0e12" : "#f0eee9"

                            Row {
                                anchors.left: parent.left
                                anchors.leftMargin: 6
                                anchors.verticalCenter: parent.verticalCenter
                                spacing: 3

                                Rectangle { width: 5; height: 5; radius: 2.5; color: "#e2342a" }
                                Rectangle { width: 5; height: 5; radius: 2.5; color: "#e67e22" }
                                Rectangle { width: 5; height: 5; radius: 2.5; color: "#2ecc71" }
                            }

                            Text {
                                anchors.centerIn: parent
                                text: "asura ~ terminal"
                                color: lightDarkButtonRoot.dark ? "#7a7a85" : "#888880"
                                font.pixelSize: 8
                                font.family: Appearance.font.family.main
                            }
                        }

                        // Terminal Content
                        Column {
                            anchors.top: winTitlebar.bottom
                            anchors.left: parent.left
                            anchors.right: parent.right
                            anchors.margins: 6
                            spacing: 3

                            Row {
                                spacing: 4
                                Text {
                                    text: "❯"
                                    color: "#e2342a"
                                    font.pixelSize: 9
                                    font.weight: Font.Bold
                                }
                                Text {
                                    text: "asura status"
                                    color: lightDarkButtonRoot.dark ? "#ffffff" : "#111111"
                                    font.pixelSize: 8
                                    font.family: Appearance.font.family.main
                                }
                            }

                            Text {
                                text: "[ACTIVE] Wayland / Hyprland"
                                color: "#2ecc71"
                                font.pixelSize: 8
                                font.family: Appearance.font.family.main
                            }

                            Row {
                                spacing: 4
                                Text {
                                    text: "❯"
                                    color: "#e2342a"
                                    font.pixelSize: 9
                                    font.weight: Font.Bold
                                }
                                Rectangle {
                                    width: 5
                                    height: 9
                                    color: lightDarkButtonRoot.dark ? "#ffffff" : "#111111"
                                    anchors.verticalCenter: parent.verticalCenter
                                }
                            }
                        }
                    }
                }
            }

            // Mode Label
            RowLayout {
                Layout.alignment: Qt.AlignHCenter
                spacing: 6
                MaterialSymbol {
                    visible: lightDarkButtonRoot.toggled
                    text: "check"
                    iconSize: 14
                    color: Appearance.m3colors.m3primary
                    Layout.alignment: Qt.AlignVCenter
                }
                StyledText {
                    text: dark ? Translation.tr("Dark Mode") : Translation.tr("Light Mode")
                    color: lightDarkButtonRoot.toggled ? Appearance.m3colors.m3primary : Appearance.colors.colOnLayer2
                    font.weight: lightDarkButtonRoot.toggled ? Font.Bold : Font.Normal
                    horizontalAlignment: Text.AlignHCenter
                    Layout.alignment: Qt.AlignVCenter
                }
            }
        }
    }
}
