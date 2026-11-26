pragma ComponentBehavior: Bound
import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import Quickshell.Hyprland

Scope {
    id: root

    PanelWindow {
        id: panelWindow
        visible: GlobalStates.calendarOpen

        WlrLayershell.namespace: "quickshell:calendarPopup"
        WlrLayershell.layer: WlrLayer.Top
        WlrLayershell.keyboardFocus: GlobalStates.calendarOpen ? WlrKeyboardFocus.OnDemand : WlrKeyboardFocus.None
        color: "transparent"

        mask: Region {
            item: GlobalStates.calendarOpen ? popupCard : null
        }

        anchors {
            top: true
            bottom: true
            left: true
            right: true
        }

        onVisibleChanged: {
            if (visible) {
                GlobalFocusGrab.addDismissable(panelWindow);
            } else {
                GlobalFocusGrab.removeDismissable(panelWindow);
            }
        }

        Connections {
            target: GlobalFocusGrab
            function onDismissed() {
                GlobalStates.calendarOpen = false;
            }
        }

        // Fullscreen backdrop click-to-dismiss
        MouseArea {
            anchors.fill: parent
            onClicked: {
                GlobalStates.calendarOpen = false;
            }
        }

        // The popup card itself
        Item {
            id: popupCardWrapper
            readonly property real barGap: 6
            anchors {
                horizontalCenter: !Config.options.bar.vertical ? parent.horizontalCenter : undefined
                left: Config.options.bar.vertical && !Config.options.bar.bottom ? parent.left : undefined
                right: Config.options.bar.vertical && Config.options.bar.bottom ? parent.right : undefined
                top: !Config.options.bar.bottom ? parent.top : undefined
                bottom: Config.options.bar.bottom && !Config.options.bar.vertical ? parent.bottom : undefined
                topMargin: !Config.options.bar.bottom ? (barGap - Appearance.sizes.elevationMargin) : 0
                bottomMargin: (Config.options.bar.bottom && !Config.options.bar.vertical) ? (barGap - Appearance.sizes.elevationMargin) : 0
                leftMargin: (Config.options.bar.vertical && !Config.options.bar.bottom) ? (barGap - Appearance.sizes.elevationMargin) : 0
                rightMargin: (Config.options.bar.vertical && Config.options.bar.bottom) ? (barGap - Appearance.sizes.elevationMargin) : 0
            }
            implicitWidth: popupCard.implicitWidth + Appearance.sizes.elevationMargin * 2
            implicitHeight: popupCard.implicitHeight + Appearance.sizes.elevationMargin * 2

            StyledRectangularShadow {
                target: popupCard
            }

            Rectangle {
                id: popupCard
                anchors.centerIn: parent
                implicitWidth: 420
                implicitHeight: cardContentLayout.implicitHeight + 32
                radius: Appearance.rounding.windowRounding
                color: Appearance.colors.colLayer0
                border.width: 1
                border.color: Appearance.colors.colLayer0Border
                clip: true

                // Consume clicks within card
                MouseArea {
                    anchors.fill: parent
                    onClicked: {}
                }

                focus: GlobalStates.calendarOpen
                Keys.onPressed: event => {
                    if (event.key === Qt.Key_Escape) {
                        GlobalStates.calendarOpen = false;
                        event.accepted = true;
                    }
                }

                ColumnLayout {
                    id: cardContentLayout
                    anchors {
                        fill: parent
                        margins: 16
                    }
                    spacing: 12

                    // Header: Time, Date & Close Button
                    RowLayout {
                        Layout.fillWidth: true
                        spacing: 10

                        ColumnLayout {
                            spacing: 2
                            StyledText {
                                text: DateTime.time
                                font.pixelSize: Appearance.font.pixelSize.huge
                                font.weight: Font.Bold
                                color: Appearance.colors.colPrimary
                            }
                            StyledText {
                                text: Qt.locale().toString(DateTime.clock.date, "dddd, MMMM dd, yyyy")
                                font.pixelSize: Appearance.font.pixelSize.small
                                color: Appearance.colors.colSubtext
                            }
                        }

                        Item {
                            Layout.fillWidth: true
                            Layout.fillHeight: false
                        }

                        RippleButton {
                            implicitWidth: 30
                            implicitHeight: 30
                            buttonRadius: Appearance.rounding.full
                            colBackground: Appearance.colors.colLayer1
                            colBackgroundHover: Appearance.colors.colLayer1Hover
                            onClicked: GlobalStates.calendarOpen = false

                            contentItem: MaterialSymbol {
                                anchors.centerIn: parent
                                text: "close"
                                iconSize: 16
                                color: Appearance.colors.colOnLayer1
                            }
                        }
                    }

                    // Segmented tab bar
                    Rectangle {
                        id: tabsRowLayout
                        Layout.fillWidth: true
                        implicitHeight: 38
                        radius: Appearance.rounding.normal
                        color: Appearance.colors.colLayer1
                        border.width: 1
                        border.color: Appearance.colors.colOutlineVariant

                        property int selectedTab: 0
                        property var tabs: [
                            {
                                "type": "calendar",
                                "name": Translation.tr("Calendar"),
                                "icon": "calendar_month",
                                "widget": Quickshell.shellPath("modules/ii/sidebarRight/calendar/CalendarWidget.qml")
                            },
                            {
                                "type": "todo",
                                "name": Translation.tr("To Do"),
                                "icon": "done_outline",
                                "widget": Quickshell.shellPath("modules/ii/sidebarRight/todo/TodoWidget.qml")
                            },
                            {
                                "type": "timer",
                                "name": Translation.tr("Timer"),
                                "icon": "schedule",
                                "widget": Quickshell.shellPath("modules/ii/sidebarRight/pomodoro/PomodoroWidget.qml")
                            }
                        ]

                        RowLayout {
                            anchors.fill: parent
                            anchors.margins: 4
                            spacing: 4

                            Repeater {
                                model: tabsRowLayout.tabs
                                RippleButton {
                                    required property int index
                                    required property var modelData
                                    readonly property bool active: tabsRowLayout.selectedTab === index
                                    Layout.fillWidth: true
                                    Layout.fillHeight: true
                                    buttonRadius: Appearance.rounding.small
                                    colBackground: active ? Appearance.colors.colSecondaryContainer : "transparent"
                                    colBackgroundHover: active ? Appearance.colors.colSecondaryContainerHover : Appearance.colors.colLayer2Hover
                                    onClicked: tabsRowLayout.selectedTab = index

                                    contentItem: RowLayout {
                                        anchors.centerIn: parent
                                        spacing: 6
                                        MaterialSymbol {
                                            text: modelData.icon
                                            iconSize: 16
                                            fill: active ? 1 : 0
                                            color: active ? Appearance.colors.colOnSecondaryContainer : Appearance.colors.colOnLayer1
                                        }
                                        StyledText {
                                            text: modelData.name
                                            font.pixelSize: Appearance.font.pixelSize.smaller
                                            font.weight: active ? Font.DemiBold : Font.Medium
                                            color: active ? Appearance.colors.colOnSecondaryContainer : Appearance.colors.colOnLayer1
                                        }
                                    }
                                }
                            }
                        }
                    }

                    // Framed content card
                    Rectangle {
                        Layout.fillWidth: true
                        Layout.preferredHeight: 332
                        radius: Appearance.rounding.normal
                        color: Appearance.colors.colLayer1
                        border.width: 1
                        border.color: Appearance.colors.colOutlineVariant
                        clip: true

                        Loader {
                            id: widgetLoader
                            anchors.fill: parent
                            anchors.margins: 10
                            source: tabsRowLayout.tabs[tabsRowLayout.selectedTab].widget
                        }
                    }
                }
            }
        }
    }

    IpcHandler {
        target: "calendar"

        function toggle(): void {
            GlobalStates.calendarOpen = !GlobalStates.calendarOpen;
        }

        function close(): void {
            GlobalStates.calendarOpen = false;
        }

        function open(): void {
            GlobalStates.calendarOpen = true;
        }
    }

    GlobalShortcut {
        name: "calendarToggle"
        description: "Toggles calendar on press"

        onPressed: {
            GlobalStates.calendarOpen = !GlobalStates.calendarOpen;
        }
    }
}
