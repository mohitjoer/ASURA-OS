import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import Quickshell
import Quickshell.Bluetooth
import Quickshell.Hyprland
import Quickshell.Io

import qs.modules.ii.sidebarRight.quickToggles
import qs.modules.ii.sidebarRight.quickToggles.classicStyle

import qs.modules.ii.sidebarRight.bluetoothDevices
import qs.modules.ii.sidebarRight.nightLight
import qs.modules.ii.sidebarRight.volumeMixer
import qs.modules.ii.sidebarRight.wifiNetworks

Item {
    id: root
    property int sidebarWidth: Appearance.sizes.sidebarWidth
    property int sidebarPadding: 8
    property bool showAudioOutputDialog: false
    property bool showAudioInputDialog: false
    property bool showBluetoothDialog: false
    property bool showNightLightDialog: false
    property bool showWifiDialog: false
    property bool editMode: false

    Connections {
        target: GlobalStates
        function onSidebarRightOpenChanged() {
            if (!GlobalStates.sidebarRightOpen) {
                root.showWifiDialog = false;
                root.showBluetoothDialog = false;
                root.showAudioOutputDialog = false;
                root.showAudioInputDialog = false;
            }
        }
    }

    implicitHeight: sidebarRightBackground.implicitHeight
    implicitWidth: sidebarRightBackground.implicitWidth

    StyledRectangularShadow {
        target: sidebarRightBackground
    }
    Rectangle {
        id: sidebarRightBackground

        anchors.fill: parent
        implicitHeight: parent.height - Appearance.sizes.hyprlandGapsOut * 2
        implicitWidth: sidebarWidth - Appearance.sizes.hyprlandGapsOut * 2
        color: Appearance.colors.colLayer0
        border.width: 1
        border.color: Appearance.colors.colLayer0Border
        radius: Appearance.rounding.screenRounding - Appearance.sizes.hyprlandGapsOut + 1

        ColumnLayout {
            anchors.fill: parent
            anchors.margins: sidebarPadding
            spacing: sidebarPadding

            SystemButtonRow {
                Layout.fillHeight: false
                Layout.fillWidth: true
                // Layout.margins: 10
                Layout.topMargin: 5
                Layout.bottomMargin: 0
            }

            Loader {
                id: slidersLoader
                Layout.fillWidth: true
                visible: active
                active: {
                    const configQuickSliders = Config.options.sidebar.quickSliders
                    if (!configQuickSliders.enable) return false
                    if (!configQuickSliders.showMic && !configQuickSliders.showVolume && !configQuickSliders.showBrightness) return false;
                    return true;
                }
                sourceComponent: QuickSliders {}
            }

            LoaderedQuickPanelImplementation {
                styleName: "classic"
                sourceComponent: ClassicQuickPanel {}
            }

            LoaderedQuickPanelImplementation {
                styleName: "android"
                sourceComponent: AndroidQuickPanel {
                    editMode: root.editMode
                }
            }

            Rectangle {
                id: layoutCard
                Layout.fillWidth: true
                radius: Appearance.rounding.normal
                color: Appearance.colors.colLayer1
                implicitHeight: layoutCol.implicitHeight + 16

                readonly property int workspaceId: Hyprland.focusedMonitor?.activeWorkspace?.id ?? 0
                property var layouts: ({})
                readonly property string current: layoutCard.layouts[String(layoutCard.workspaceId)] ?? "default"

                FileView {
                    path: (Quickshell.env("XDG_CONFIG_HOME") || (Quickshell.env("HOME") + "/.config")) + "/toren/layouts.json"
                    watchChanges: true
                    printErrors: false
                    onFileChanged: reload()
                    onLoaded: {
                        try {
                            layoutCard.layouts = JSON.parse(text()).workspaces ?? {};
                        } catch (e) {
                            layoutCard.layouts = {};
                        }
                    }
                    onLoadFailed: layoutCard.layouts = {}
                }

                ColumnLayout {
                    id: layoutCol
                    anchors.fill: parent
                    anchors.margins: 8
                    spacing: 6

                    RowLayout {
                        Layout.fillWidth: true
                        spacing: 6

                        MaterialSymbol {
                            text: {
                                switch (layoutCard.current) {
                                    case "tile": return "dashboard";
                                    case "scroll": return "view_week";
                                    case "master": return "view_quilt";
                                    case "monocle": return "crop_square";
                                    case "float": return "filter_none";
                                    default: return "dashboard";
                                }
                            }
                            iconSize: 18
                            color: Appearance.colors.colPrimary
                        }

                        StyledText {
                            Layout.fillWidth: true
                            text: Translation.tr("Workspace %1 Layout").arg(layoutCard.workspaceId)
                            font.pixelSize: Appearance.font.pixelSize.small
                            font.weight: Font.DemiBold
                            color: Appearance.colors.colOnLayer1
                        }

                        StyledText {
                            text: {
                                switch (layoutCard.current) {
                                    case "tile": return Translation.tr("Tiling");
                                    case "scroll": return Translation.tr("Scrolling");
                                    case "master": return Translation.tr("Master");
                                    case "monocle": return Translation.tr("Monocle");
                                    case "float": return Translation.tr("Floating");
                                    default: return Translation.tr("Default");
                                }
                            }
                            font.pixelSize: Appearance.font.pixelSize.smaller
                            color: Appearance.colors.colSubtext
                        }
                    }

                    RowLayout {
                        Layout.fillWidth: true
                        spacing: 4

                        Repeater {
                            model: [
                                { id: "tile", icon: "dashboard", label: Translation.tr("Tile") },
                                { id: "scroll", icon: "view_week", label: Translation.tr("Scroll") },
                                { id: "master", icon: "view_quilt", label: Translation.tr("Master") },
                                { id: "monocle", icon: "crop_square", label: Translation.tr("Monocle") },
                                { id: "float", icon: "filter_none", label: Translation.tr("Float") }
                            ]
                            delegate: Rectangle {
                                id: btn
                                required property var modelData
                                required property int index
                                Layout.fillWidth: true
                                implicitHeight: 32
                                radius: Appearance.rounding.small
                                readonly property bool isSelected: layoutCard.current === modelData.id || (layoutCard.current === "default" && modelData.id === "tile")

                                color: isSelected
                                    ? Appearance.colors.colPrimary
                                    : (btnMouse.containsMouse ? Appearance.colors.colLayer1Hover : Appearance.colors.colSecondaryContainer)

                                Behavior on color {
                                    animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
                                }

                                RowLayout {
                                    anchors.centerIn: parent
                                    spacing: 3
                                    MaterialSymbol {
                                        text: modelData.icon
                                        iconSize: 15
                                        color: btn.isSelected ? Appearance.colors.colOnPrimary : Appearance.colors.colOnSecondaryContainer
                                    }
                                    StyledText {
                                        text: modelData.label
                                        font.pixelSize: Appearance.font.pixelSize.smaller
                                        font.weight: btn.isSelected ? Font.Bold : Font.Normal
                                        color: btn.isSelected ? Appearance.colors.colOnPrimary : Appearance.colors.colOnSecondaryContainer
                                    }
                                }

                                MouseArea {
                                    id: btnMouse
                                    anchors.fill: parent
                                    hoverEnabled: true
                                    cursorShape: Qt.PointingHandCursor
                                    onClicked: {
                                        Quickshell.execDetached(["toren", "layout", "set", modelData.id]);
                                    }
                                }

                                StyledToolTip {
                                    extraVisibleCondition: btnMouse.containsMouse
                                    text: modelData.label
                                }
                            }
                        }
                    }
                }
            }

            CenterWidgetGroup {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillHeight: true
                Layout.fillWidth: true
            }
        }
    }

    ToggleDialog {
        shownPropertyString: "showAudioOutputDialog"
        dialog: VolumeDialog {
            isSink: true
        }
    }

    ToggleDialog {
        shownPropertyString: "showAudioInputDialog"
        dialog: VolumeDialog {
            isSink: false
        }
    }

    ToggleDialog {
        shownPropertyString: "showBluetoothDialog"
        dialog: BluetoothDialog {}
        onShownChanged: {
            if (!shown) {
                Bluetooth.defaultAdapter.discovering = false;
            } else {
                Bluetooth.defaultAdapter.enabled = true;
                Bluetooth.defaultAdapter.discovering = true;
            }
        }
    }

    ToggleDialog {
        shownPropertyString: "showNightLightDialog"
        dialog: NightLightDialog {}
    }

    ToggleDialog {
        shownPropertyString: "showWifiDialog"
        dialog: WifiDialog {}
        onShownChanged: {
            if (!shown) return;
            Network.enableWifi();
            Network.rescanWifi();
        }
    }

    component ToggleDialog: Loader {
        id: toggleDialogLoader
        required property string shownPropertyString
        property alias dialog: toggleDialogLoader.sourceComponent
        readonly property bool shown: root[shownPropertyString]
        anchors.fill: parent

        onShownChanged: if (shown) toggleDialogLoader.active = true;
        active: shown
        onActiveChanged: {
            if (active) {
                item.show = true;
                item.forceActiveFocus();
            }
        }
        Connections {
            target: toggleDialogLoader.item
            function onDismiss() {
                toggleDialogLoader.item.show = false
                root[toggleDialogLoader.shownPropertyString] = false;
            }
            function onVisibleChanged() {
                if (!toggleDialogLoader.item.visible && !root[toggleDialogLoader.shownPropertyString]) toggleDialogLoader.active = false;
            }
        }
    }

    component LoaderedQuickPanelImplementation: Loader {
        id: quickPanelImplLoader
        required property string styleName
        Layout.alignment: item?.Layout.alignment ?? Qt.AlignHCenter
        Layout.fillWidth: item?.Layout.fillWidth ?? false
        visible: active
        active: Config.options.sidebar.quickToggles.style === styleName
        Connections {
            target: quickPanelImplLoader.item
            function onOpenAudioOutputDialog() {
                root.showAudioOutputDialog = true;
            }
            function onOpenAudioInputDialog() {
                root.showAudioInputDialog = true;
            }
            function onOpenBluetoothDialog() {
                root.showBluetoothDialog = true;
            }
            function onOpenNightLightDialog() {
                root.showNightLightDialog = true;
            }
            function onOpenWifiDialog() {
                root.showWifiDialog = true;
            }
        }
    }

    component SystemButtonRow: Item {
        implicitHeight: 32

        Rectangle {
            id: uptimeContainer
            anchors {
                top: parent.top
                bottom: parent.bottom
                left: parent.left
            }
            color: Appearance.colors.colLayer1
            radius: height / 2
            implicitWidth: uptimeRow.implicitWidth + 16
            implicitHeight: 32
            
            Row {
                id: uptimeRow
                anchors.centerIn: parent
                spacing: 6
                CustomIcon {
                    id: distroIcon
                    anchors.verticalCenter: parent.verticalCenter
                    width: 18
                    height: 18
                    source: SystemInfo.distroIcon
                    colorize: true
                    color: Appearance.colors.colOnLayer0
                }
                StyledText {
                    anchors.verticalCenter: parent.verticalCenter
                    font.pixelSize: Appearance.font.pixelSize.small
                    color: Appearance.colors.colOnLayer0
                    text: Translation.tr("Up %1").arg(DateTime.uptime)
                    textFormat: Text.MarkdownText
                }
            }
        }

        ButtonGroup {
            id: systemButtonsRow
            anchors {
                top: parent.top
                bottom: parent.bottom
                right: parent.right
            }
            color: Appearance.colors.colLayer1
            padding: 2
            implicitHeight: 32

            QuickToggleButton {
                toggled: root.editMode
                visible: Config.options.sidebar.quickToggles.style === "android"
                buttonIcon: "edit"
                onClicked: root.editMode = !root.editMode
                StyledToolTip {
                    text: Translation.tr("Edit quick toggles") + (root.editMode ? Translation.tr("\nLMB to enable/disable\nRMB to toggle size\nScroll to swap position") : "")
                }
            }
            QuickToggleButton {
                toggled: false
                buttonIcon: "restart_alt"
                onClicked: {
                    Quickshell.execDetached(["hyprctl", "reload"])
                    Quickshell.reload(true);
                }
                StyledToolTip {
                    text: Translation.tr("Reload Hyprland & Quickshell")
                }
            }
            QuickToggleButton {
                toggled: false
                buttonIcon: "settings"
                onClicked: {
                    GlobalStates.sidebarRightOpen = false;
                    // Toren has one settings panel: the Hub, on the desktop's page.
                    Quickshell.execDetached(["toren-shell", "hub", "open", "desktop-shell"]);
                }
                StyledToolTip {
                    text: Translation.tr("Settings")
                }
            }
            QuickToggleButton {
                toggled: false
                buttonIcon: "power_settings_new"
                onClicked: {
                    GlobalStates.sessionOpen = true;
                }
                StyledToolTip {
                    text: Translation.tr("Session")
                }
            }
        }
    }
}
