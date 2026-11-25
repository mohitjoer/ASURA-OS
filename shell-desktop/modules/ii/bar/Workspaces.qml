pragma ComponentBehavior: Bound
import qs
import qs.services
import qs.modules.common
import qs.modules.common.widgets
import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Hyprland
import Quickshell.Widgets

// ASURA overlay (shell-desktop/toren/vendor.sh OVERLAYS): the bar's workspaces,
// showing only the workspaces that exist, the active one and those with
// windows, in order. end-4's widget draws a fixed row of slots
// (bar.workspaces.shown, 10) and marks the empty ones with dots; Hyprland
// itself keeps no empty workspace but the active one, so its list is exactly
// what to show. Each workspace shows its main window's icon, or a dot when it
// has none; the active one is underlined. Left click switches, the wheel
// steps through workspaces, right click opens the overview (BarContent).
ButtonMouseArea {
    id: root

    readonly property HyprlandMonitor monitor: Hyprland.monitorFor(root.QsWindow.window?.screen)
    readonly property int activeId: root.monitor?.activeWorkspace?.id ?? 1
    property bool vertical: Config.options.bar.vertical
    readonly property real barThickness: vertical ? Appearance.sizes.verticalBarWidth : Appearance.sizes.barHeight
    // BarContent pads its group with this (end-4 reads it but never defined it).
    property real widgetPadding: 4
    property real workspaceButtonWidth: 26
    property real iconSize: workspaceButtonWidth * 0.69

    // The workspaces on this monitor: every regular one Hyprland has with windows,
    // plus the active one (at most one empty workspace is shown).
    readonly property var shownIds: {
        const monitorId = root.monitor?.id;
        const withWindows = HyprlandData.workspaces
            .filter(ws => ws.id > 0 && (monitorId === undefined || ws.monitorID === monitorId) && ws.windows > 0)
            .map(ws => ws.id);

        let ids = [...withWindows];
        if (root.activeId > 0 && !ids.includes(root.activeId)) {
            ids.push(root.activeId);
        }
        return ids.sort((a, b) => a - b);
    }

    Layout.alignment: vertical ? Qt.AlignHCenter : Qt.AlignVCenter
    Layout.fillWidth: vertical
    Layout.fillHeight: !vertical
    implicitWidth: vertical ? barThickness : row.implicitWidth
    implicitHeight: vertical ? row.implicitHeight : barThickness

    // Clicks belong to each workspace's own MouseArea; this one takes the wheel.
    acceptedButtons: Qt.NoButton
    hoverEnabled: true
    onWheel: event => {
        if (event.angleDelta.y < 0)
            Hyprland.dispatch("workspace r+1");
        else if (event.angleDelta.y > 0)
            Hyprland.dispatch("workspace r-1");
    }

    GridLayout {
        id: row
        anchors.centerIn: parent
        flow: root.vertical ? GridLayout.TopToBottom : GridLayout.LeftToRight
        rowSpacing: 2
        columnSpacing: 2

        // ASURA Logo Button
        Rectangle {
            id: torenLogoBtn
            readonly property bool hovered: logoMouse.containsMouse
            implicitWidth: root.workspaceButtonWidth + 2
            implicitHeight: root.workspaceButtonWidth
            radius: 4
            color: logoMouse.containsMouse ? Appearance.colors.colLayer1Hover : "transparent"
            Behavior on color {
                animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
            }

            Image {
                anchors.centerIn: parent
                width: 18
                height: 18
                source: Appearance.m3colors.darkmode 
                    ? Quickshell.shellPath("assets/icons/toren-logo.svg") 
                    : Quickshell.shellPath("assets/icons/toren-logo-dark.svg")
                fillMode: Image.PreserveAspectFit
                smooth: true
            }

            MouseArea {
                id: logoMouse
                anchors.fill: parent
                hoverEnabled: true
                cursorShape: Qt.PointingHandCursor
                acceptedButtons: Qt.LeftButton | Qt.RightButton
                onClicked: event => {
                    if (event.button === Qt.RightButton) {
                        Quickshell.execDetached(["toren-shell", "hub", "open"]);
                    } else {
                        GlobalStates.overviewOpen = !GlobalStates.overviewOpen;
                    }
                }
            }

            StyledToolTip {
                extraVisibleCondition: logoMouse.containsMouse
                text: Translation.tr("TOREN Menu (Super)")
            }
        }

        Repeater {
            model: root.shownIds

            delegate: Rectangle {
                id: ws
                required property int modelData
                readonly property bool active: modelData === root.activeId
                readonly property var mainWindow: HyprlandData.biggestWindowForWorkspace(modelData)
                readonly property bool hovered: wsMouse.containsMouse

                implicitWidth: root.workspaceButtonWidth
                implicitHeight: root.workspaceButtonWidth
                radius: 4
                // No fill for the active one: it is underlined. Hover tints.
                color: ws.active ? Appearance.colors.colSecondaryContainer : (ws.hovered ? Appearance.colors.colLayer1Hover : "transparent")
                Behavior on color {
                    animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
                }

                IconImage {
                    id: wsIcon
                    anchors.centerIn: parent
                    visible: !!ws.mainWindow && (Config.options?.bar.workspaces.showAppIcons ?? true) && status === Image.Ready
                    implicitSize: root.iconSize
                    source: {
                        if (!ws.mainWindow?.class) return "";
                        const cls = ws.mainWindow.class;
                        const icon = cls === "com.mitchellh.ghostty" || cls === "ghostty" ? "com.mitchellh.ghostty" : AppSearch.guessIcon(cls);
                        const path = Quickshell.iconPath(icon);
                        return path ? path : (Quickshell.iconPath("terminal") || "");
                    }
                    opacity: ws.active ? 1 : 0.75
                }
                StyledText {
                    anchors.centerIn: parent
                    visible: !wsIcon.visible
                    text: `${ws.modelData}`
                    font.pixelSize: Appearance.font.pixelSize.small
                    font.weight: ws.active ? Font.Bold : Font.DemiBold
                    color: ws.active ? Appearance.colors.colPrimary : (ws.hovered ? Appearance.colors.colOnLayer1 : Appearance.colors.colSubtext)
                }
                Rectangle { // the active workspace's underline
                    anchors {
                        bottom: parent.bottom
                        horizontalCenter: parent.horizontalCenter
                    }
                    width: ws.active ? parent.width - 6 : 0
                    height: 2
                    color: Appearance.colors.colPrimary
                    Behavior on width {
                        animation: Appearance.animation.elementMoveFast.numberAnimation.createObject(this)
                    }
                }

                MouseArea {
                    id: wsMouse
                    anchors.fill: parent
                    hoverEnabled: true
                    acceptedButtons: Qt.LeftButton
                    cursorShape: Qt.PointingHandCursor
                    onClicked: Hyprland.dispatch(`workspace ${ws.modelData}`)
                }

                StyledToolTip {
                    extraVisibleCondition: wsMouse.containsMouse
                    text: ws.mainWindow ? `${ws.modelData}: ${ws.mainWindow.title || ws.mainWindow.class}` : Translation.tr("Workspace %1").arg(ws.modelData)
                }
            }
        }

        // Add tab (+) button
        Rectangle {
            id: addTabButton
            readonly property bool hovered: addTabMouse.containsMouse
            implicitWidth: root.workspaceButtonWidth
            implicitHeight: root.workspaceButtonWidth
            radius: 4
            color: addTabMouse.containsMouse ? Appearance.colors.colLayer1Hover : "transparent"
            Behavior on color {
                animation: Appearance.animation.elementMoveFast.colorAnimation.createObject(this)
            }

            MaterialSymbol {
                anchors.centerIn: parent
                text: "add"
                iconSize: 16
                color: addTabMouse.containsMouse ? Appearance.colors.colOnLayer1 : Appearance.colors.colSubtext
            }

            MouseArea {
                id: addTabMouse
                anchors.fill: parent
                hoverEnabled: true
                cursorShape: Qt.PointingHandCursor
                acceptedButtons: Qt.LeftButton
                onClicked: {
                    let maxId = 0;
                    for (let id of root.shownIds) {
                        if (id > maxId) maxId = id;
                    }
                    Hyprland.dispatch(`workspace ${maxId + 1}`);
                }
            }

            StyledToolTip {
                extraVisibleCondition: addTabMouse.containsMouse
                text: Translation.tr("New workspace")
            }
        }
    }
}
