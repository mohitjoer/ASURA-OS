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
            anchors {
                horizontalCenter: !Config.options.bar.vertical ? parent.horizontalCenter : undefined
                left: Config.options.bar.vertical && !Config.options.bar.bottom ? parent.left : undefined
                right: Config.options.bar.vertical && Config.options.bar.bottom ? parent.right : undefined
                top: !Config.options.bar.bottom ? parent.top : undefined
                bottom: Config.options.bar.bottom && !Config.options.bar.vertical ? parent.bottom : undefined
                topMargin: !Config.options.bar.bottom ? (Appearance.sizes.barHeight + Appearance.sizes.hyprlandGapsOut) : 0
                bottomMargin: (Config.options.bar.bottom && !Config.options.bar.vertical) ? (Appearance.sizes.barHeight + Appearance.sizes.hyprlandGapsOut) : 0
                leftMargin: (Config.options.bar.vertical && !Config.options.bar.bottom) ? (Appearance.sizes.verticalBarWidth + Appearance.sizes.hyprlandGapsOut) : 0
                rightMargin: (Config.options.bar.vertical && Config.options.bar.bottom) ? (Appearance.sizes.verticalBarWidth + Appearance.sizes.hyprlandGapsOut) : 0
            }
            implicitWidth: popupCard.implicitWidth + Appearance.sizes.elevationMargin * 2
            implicitHeight: popupCard.implicitHeight + Appearance.sizes.elevationMargin * 2

            StyledRectangularShadow {
                target: popupCard
            }

            Rectangle {
                id: popupCard
                anchors.centerIn: parent
                implicitWidth: 430
                implicitHeight: cardContentLayout.implicitHeight + 28
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
                                font.pixelSize: Appearance.font.pixelSize.normal
                                color: Appearance.colors.colOnLayer0
                            }
                        }

                        Item {
                            Layout.fillWidth: true
                            Layout.fillHeight: false
                        }

                        RippleButton {
                            implicitWidth: 32
                            implicitHeight: 32
                            buttonRadius: Appearance.rounding.full
                            colBackground: Appearance.colors.colLayer1
                            colBackgroundHover: Appearance.colors.colLayer1Hover
                            onClicked: GlobalStates.calendarOpen = false

                            contentItem: MaterialSymbol {
                                anchors.centerIn: parent
                                text: "close"
                                iconSize: 18
                                color: Appearance.colors.colOnLayer1
                            }
                        }
                    }

                    // Divider
                    Rectangle {
                        Layout.fillWidth: true
                        implicitHeight: 1
                        color: Appearance.colors.colOutlineVariant
                    }

                    // Main Content: Tabs Rail + Content Loader
                    RowLayout {
                        id: tabsRowLayout
                        Layout.fillWidth: true
                        Layout.preferredHeight: 340
                        spacing: 14

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

                        // Navigation rail
                        Item {
                            Layout.fillHeight: true
                            implicitWidth: tabRail.implicitWidth

                            NavigationRailTabArray {
                                id: tabRail
                                anchors.verticalCenter: parent.verticalCenter
                                anchors.left: parent.left
                                currentIndex: tabsRowLayout.selectedTab
                                expanded: false

                                Repeater {
                                    model: tabsRowLayout.tabs
                                    NavigationRailButton {
                                        required property int index
                                        required property var modelData
                                        showToggledHighlight: false
                                        toggled: tabsRowLayout.selectedTab === index
                                        buttonText: modelData.name
                                        buttonIcon: modelData.icon
                                        onPressed: {
                                            tabsRowLayout.selectedTab = index;
                                        }
                                    }
                                }
                            }
                        }

                        // Content View
                        Item {
                            Layout.fillWidth: true
                            Layout.fillHeight: true

                            Loader {
                                id: widgetLoader
                                anchors.fill: parent
                                source: tabsRowLayout.tabs[tabsRowLayout.selectedTab].widget
                            }
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
