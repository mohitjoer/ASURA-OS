import qs.modules.ii.bar.weather
import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Services.UPower
import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.functions

// ASURA overlay (shell-desktop/asura/vendor.sh OVERLAYS): the bar's layout.
// Left: the workspaces, then media while a player has a track (no active
// window title).
// Centre: the clock and date, alone. Right, from the end: the battery, the
// resources, the status indicators (they open the control menu),
// the tray, weather. end-4's util buttons (screen snip, keyboard, brightness)
// are not on the bar. Only the buttons open the sidebars.
Item { // Bar content region
    id: root

    property var screen: root.QsWindow.window?.screen
    property var brightnessMonitor: Brightness.getMonitorForScreen(screen)
    property real useShortenedForm: (Appearance.sizes.barHellaShortenScreenWidthThreshold >= screen?.width) ? 2 : (Appearance.sizes.barShortenScreenWidthThreshold >= screen?.width) ? 1 : 0
    readonly property int centerSideModuleWidth: (useShortenedForm == 2) ? Appearance.sizes.barCenterSideModuleWidthHellaShortened : (useShortenedForm == 1) ? Appearance.sizes.barCenterSideModuleWidthShortened : Appearance.sizes.barCenterSideModuleWidth

    component VerticalBarSeparator: Rectangle {
        Layout.topMargin: Appearance.sizes.baseBarHeight / 3
        Layout.bottomMargin: Appearance.sizes.baseBarHeight / 3
        Layout.fillHeight: true
        implicitWidth: 1
        color: Appearance.colors.colOutlineVariant
    }

    // Background shadow
    Loader {
        active: Config.options.bar.showBackground && Config.options.bar.cornerStyle === 1 && Config.options.bar.floatStyleShadow
        anchors.fill: barBackground
        sourceComponent: StyledRectangularShadow {
            anchors.fill: undefined // The loader's anchors act on this, and this should not have any anchor
            target: barBackground
        }
    }
    // Background
    Rectangle {
        id: barBackground
        anchors {
            fill: parent
            margins: Config.options.bar.cornerStyle === 1 ? (Appearance.sizes.hyprlandGapsOut) : 0 // idk why but +1 is needed
        }
        color: Config.options.bar.showBackground ? Appearance.colors.colLayer0 : "transparent"
        radius: Config.options.bar.cornerStyle === 1 ? Appearance.rounding.windowRounding : 0
        border.width: Config.options.bar.cornerStyle === 1 ? 1 : 0
        border.color: Appearance.colors.colLayer0Border
    }

    FocusedScrollMouseArea { // Left side | scroll to change brightness
        id: barLeftSideMouseArea

        anchors {
            top: parent.top
            bottom: parent.bottom
            left: parent.left
            right: middleSection.left
        }
        implicitWidth: leftSectionRowLayout.implicitWidth
        implicitHeight: Appearance.sizes.baseBarHeight

        // No scroll or click action: volume and brightness are sliders in the
        // control menu (the right sidebar), and only its button opens it.

        // Visual content

        RowLayout {
            id: leftSectionRowLayout
            anchors.fill: parent
            spacing: 0

            BarGroup { // Workspaces, at the start of the bar
                id: workspacesGroup
                Layout.alignment: Qt.AlignVCenter
                Layout.leftMargin: 8
                padding: workspacesWidget.widgetPadding

                Workspaces {
                    id: workspacesWidget
                    Layout.fillHeight: true
                    MouseArea {
                        // Right-click to toggle overview
                        anchors.fill: parent
                        acceptedButtons: Qt.RightButton

                        onPressed: event => {
                            if (event.button === Qt.RightButton) {
                                GlobalStates.overviewOpen = !GlobalStates.overviewOpen;
                            }
                        }
                    }
                }
            }

            // Media, second from the start, while a player has a track or is playing
            BarGroup {
                id: mediaGroup
                visible: !!MprisController.activePlayer && (
                    MprisController.activePlayer.isPlaying ||
                    (MprisController.activePlayer.trackTitle ?? "").trim() !== ""
                )
                Layout.alignment: Qt.AlignVCenter
                Layout.leftMargin: 6
                padding: 4

                Media {
                    Layout.fillHeight: true
                }
            }

            Item {
                Layout.fillWidth: true
            }
        }
    }

    Row { // Middle section
        id: middleSection
        anchors {
            top: parent.top
            bottom: parent.bottom
            horizontalCenter: parent.horizontalCenter
        }
        spacing: 4

        MouseArea {
            id: rightCenterGroup
            anchors.verticalCenter: parent.verticalCenter
            implicitWidth: rightCenterGroupContent.implicitWidth
            implicitHeight: rightCenterGroupContent.implicitHeight
            cursorShape: Qt.PointingHandCursor

            onClicked: {
                GlobalStates.calendarOpen = !GlobalStates.calendarOpen;
            }

            BarGroup {
                id: rightCenterGroupContent
                anchors.fill: parent

                ClockWidget {
                    showDate: (Config.options.bar.verbose && root.useShortenedForm < 2)
                    Layout.alignment: Qt.AlignVCenter
                    Layout.fillWidth: true
                }

            }
        }
    }

    FocusedScrollMouseArea { // Right side | scroll to change volume
        id: barRightSideMouseArea

        anchors {
            top: parent.top
            bottom: parent.bottom
            left: middleSection.right
            right: parent.right
        }
        implicitWidth: rightSectionRowLayout.implicitWidth
        implicitHeight: Appearance.sizes.baseBarHeight

        onScrollDown: Audio.decrementVolume()
        onScrollUp: Audio.incrementVolume()

        // Visual content

        RowLayout {
            id: rightSectionRowLayout
            anchors.fill: parent
            spacing: 5
            layoutDirection: Qt.RightToLeft

            RippleButton { // Unified quick settings & status capsule, at the very end of the bar
                id: rightSidebarButton

                Layout.alignment: Qt.AlignVCenter
                Layout.rightMargin: 10
                Layout.fillWidth: false

                implicitWidth: indicatorsRowLayout.implicitWidth + 24
                implicitHeight: indicatorsRowLayout.implicitHeight + 8

                buttonRadius: Appearance.rounding.full
                colBackground: hovered ? Appearance.colors.colLayer1Hover : Appearance.colors.colLayer1
                colBackgroundHover: Appearance.colors.colLayer1Hover
                colRipple: Appearance.colors.colLayer1Active
                colBackgroundToggled: Appearance.colors.colSecondaryContainer
                colBackgroundToggledHover: Appearance.colors.colSecondaryContainerHover
                colRippleToggled: Appearance.colors.colSecondaryContainerActive
                toggled: GlobalStates.sidebarRightOpen
                property color colText: toggled ? Appearance.m3colors.m3onSecondaryContainer : Appearance.colors.colOnLayer0

                Behavior on colText {
                    animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
                }

                onClicked: {
                    GlobalStates.sidebarRightOpen = !GlobalStates.sidebarRightOpen;
                }

                WheelHandler {
                    onWheel: event => {
                        if (event.angleDelta.y > 0)
                            Audio.incrementVolume();
                        else
                            Audio.decrementVolume();
                    }
                }

                RowLayout {
                    id: indicatorsRowLayout
                    anchors.centerIn: parent
                    spacing: 10

                    MaterialSymbol { // Network
                        Layout.alignment: Qt.AlignVCenter
                        text: Network.materialSymbol
                        iconSize: Appearance.font.pixelSize.larger
                        color: rightSidebarButton.colText
                    }

                    MaterialSymbol { // Bluetooth
                        Layout.alignment: Qt.AlignVCenter
                        visible: true
                        text: BluetoothStatus.connected ? "bluetooth_connected" : (BluetoothStatus.available && !BluetoothStatus.enabled) ? "bluetooth_disabled" : "bluetooth"
                        iconSize: Appearance.font.pixelSize.larger
                        color: rightSidebarButton.colText
                    }

                    MaterialSymbol { // Volume
                        Layout.alignment: Qt.AlignVCenter
                        text: Audio.sink?.audio?.muted ? "volume_off" : (Audio.sink?.audio?.volume === 0 ? "volume_mute" : (Audio.sink?.audio?.volume < 0.5 ? "volume_down" : "volume_up"))
                        iconSize: Appearance.font.pixelSize.larger
                        color: rightSidebarButton.colText
                    }

                    Revealer {
                        reveal: Audio.source?.audio?.muted ?? false
                        Layout.fillHeight: true
                        Layout.alignment: Qt.AlignVCenter
                        MaterialSymbol {
                            text: "mic_off"
                            iconSize: Appearance.font.pixelSize.larger
                            color: Appearance.colors.colError
                        }
                    }

                    HyprlandXkbIndicator {
                        Layout.alignment: Qt.AlignVCenter
                        color: rightSidebarButton.colText
                    }

                    Revealer {
                        reveal: Notifications.silent || Notifications.unread > 0
                        Layout.fillHeight: true
                        Layout.alignment: Qt.AlignVCenter
                        implicitHeight: reveal ? notificationUnreadCount.implicitHeight : 0
                        implicitWidth: reveal ? notificationUnreadCount.implicitWidth : 0
                        NotificationUnreadCount {
                            id: notificationUnreadCount
                        }
                    }

                    RowLayout { // Battery
                        id: batteryGroup
                        visible: Battery.available
                        spacing: 5
                        Layout.alignment: Qt.AlignVCenter

                        readonly property bool isCharging: Battery.isCharging
                        readonly property real percentage: Battery.percentage
                        readonly property bool isLow: percentage <= (Config.options?.battery?.low ?? 20) / 100
                        readonly property color levelColor: (isLow && !isCharging) ? Appearance.colors.colError
                            : isCharging ? Appearance.colors.colPrimary
                            : rightSidebarButton.colText

                        Item { // Battery icon
                            implicitWidth: 22
                            implicitHeight: 12
                            Layout.alignment: Qt.AlignVCenter

                            Rectangle { // body
                                id: batteryBody
                                width: 20
                                height: 12
                                color: "transparent"
                                border.width: 1.5
                                border.color: rightSidebarButton.colText
                                radius: 2

                                Rectangle { // charge
                                    x: 2.5
                                    y: 2.5
                                    height: parent.height - 5
                                    width: Math.max(1, (parent.width - 5) * Math.max(0, Math.min(1, batteryGroup.percentage)))
                                    color: batteryGroup.levelColor
                                    radius: 1
                                    Behavior on width {
                                        animation: Appearance.animation.elementMoveFast.numberAnimation.createObject(this)
                                    }
                                }
                            }

                            Rectangle { // terminal
                                anchors {
                                    left: batteryBody.right
                                    verticalCenter: batteryBody.verticalCenter
                                }
                                width: 2
                                height: 5
                                color: rightSidebarButton.colText
                            }

                            MaterialSymbol { // charging bolt
                                anchors.centerIn: batteryBody
                                visible: batteryGroup.isCharging
                                fill: 1
                                text: "bolt"
                                iconSize: 13
                                color: Appearance.colors.colLayer0
                                style: Text.Outline
                                styleColor: Appearance.colors.colLayer0
                            }
                        }

                        StyledText {
                            Layout.alignment: Qt.AlignVCenter
                            text: `${Math.round(batteryGroup.percentage * 100)}%`
                            color: (batteryGroup.isLow && !batteryGroup.isCharging) ? Appearance.colors.colError : rightSidebarButton.colText
                            font.pixelSize: Appearance.font.pixelSize.small
                            font.family: Appearance.font.family.monospace
                        }

                        BatteryPopup {
                            id: batteryPopup
                            hoverTarget: batteryGroup
                        }
                    }
                }
            }

            BarGroup { // Resources, to the left of the status pill
                Layout.alignment: Qt.AlignVCenter
                Layout.rightMargin: 4
                Resources {
                    alwaysShowAllResources: root.useShortenedForm === 2
                }
            }

            SysTray {
                visible: root.useShortenedForm === 0
                Layout.fillWidth: false
                Layout.fillHeight: true
                invertSide: Config?.options.bar.bottom
            }

            Item {
                Layout.fillWidth: true
                Layout.fillHeight: true
            }

            // Weather
            Loader {
                Layout.leftMargin: 4
                active: Config.options.bar.weather.enable

                sourceComponent: BarGroup {
                    WeatherBar {}
                }
            }
        }
    }
}
