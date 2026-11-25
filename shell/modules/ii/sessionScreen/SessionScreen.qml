// The session screen, as Toren ships it.
//
// This is an overlay. Upstream's version of this file is the same menu — a
// translucent overlay layer over the focused screen, a title, a row of square
// action buttons with keyboard navigation, a subtitle naming the focused
// button, and two warnings for a running package manager or download. The
// differences are who dispatches it and where the actions go.
//
// Upstream's entry point registers one IPC handler (target "session") and
// three global shortcuts of its own. Toren's registers none: the toggle
// reaches the flag through the shell's own IPC bridge (`toren ipc call toren
// session`, bound to Ctrl+Alt+Delete — the chord upstream binds to this
// screen's own shortcut), the sidebar's power button opens it, and a click
// anywhere, Escape, or the lock engaging are the ways out.
//
// The buttons route into Toren's session machinery rather than upstream's
// Session singleton, whose logout kills every window pid and then pkills the
// compositor and whose poweroff runs loginctl directly. Lock runs
// `toren-shell lock` — the daemon spawns hyprlock, and that is the only lock
// surface this desktop has. Logout, reboot and shutdown raise the shell's
// confirmation dialog (ShellState.askSessionAction), which runs the action
// through the daemon only on its positive press — reboot and shutdown are
// the two actions that dialog exists to confirm. Upstream's other four
// actions are not here: the daemon deliberately ships no suspend or
// hibernate, and there is no task-manager or firmware-reboot command behind
// this screen, and a button with nothing behind it is a control that runs
// nothing. The four that remain sit in the row upstream's grid lays out,
// ordered by how much they end: lock, logout, reboot, shutdown.

import qs
import qs.services
import qs.Services
import qs.modules.common
import qs.modules.common.widgets
import qs.modules.common.functions
import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Wayland
import Quickshell.Hyprland

Scope {
    id: root
    property var focusedScreen: Quickshell.screens.find(s => s.name === Hyprland.focusedMonitor?.name)

    Loader {
        id: sessionLoader
        active: GlobalStates.sessionOpen
        onActiveChanged: {
            if (sessionLoader.active)
                SessionWarnings.refresh();
        }

        Connections {
            target: GlobalStates
            function onScreenLockedChanged() {
                if (GlobalStates.screenLocked) {
                    GlobalStates.sessionOpen = false;
                }
            }
        }

        sourceComponent: PanelWindow { // Session menu
            id: sessionRoot
            visible: sessionLoader.active
            property string subtitle

            function hide() {
                GlobalStates.sessionOpen = false;
            }

            exclusionMode: ExclusionMode.Ignore
            WlrLayershell.namespace: "quickshell:session"
            WlrLayershell.layer: WlrLayer.Overlay
            WlrLayershell.keyboardFocus: WlrKeyboardFocus.Exclusive
            color: ColorUtils.transparentize(Appearance.m3colors.m3background, Appearance.m3colors.darkmode ? 0.05 : 0.12)

            anchors {
                top: true
                left: true
                right: true
            }

            implicitWidth: root.focusedScreen?.width ?? 0
            implicitHeight: root.focusedScreen?.height ?? 0

            MouseArea {
                id: sessionMouseArea
                anchors.fill: parent
                onClicked: {
                    sessionRoot.hide();
                }
            }

            ColumnLayout { // Content column
                id: contentColumn
                anchors.centerIn: parent
                spacing: 15

                Keys.onPressed: event => {
                    if (event.key === Qt.Key_Escape) {
                        sessionRoot.hide();
                    }
                }

                ColumnLayout {
                    Layout.alignment: Qt.AlignHCenter
                    spacing: 0
                    StyledText {
                        // Title
                        Layout.alignment: Qt.AlignHCenter
                        horizontalAlignment: Text.AlignHCenter
                        font {
                            family: Appearance.font.family.title
                            pixelSize: Appearance.font.pixelSize.title
                            variableAxes: Appearance.font.variableAxes.title
                        }
                        text: Translation.tr("Session")
                    }

                    StyledText {
                        // Small instruction
                        Layout.alignment: Qt.AlignHCenter
                        horizontalAlignment: Text.AlignHCenter
                        font.pixelSize: Appearance.font.pixelSize.normal
                        text: Translation.tr("Arrow keys to navigate, Enter to select\nEsc or click anywhere to cancel")
                    }
                }

                GridLayout {
                    columns: sessionHibernate.visible ? 5 : 4
                    columnSpacing: 15
                    rowSpacing: 15

                    SessionActionButton {
                        id: sessionLock
                        focus: sessionRoot.visible
                        buttonIcon: "lock"
                        buttonText: Translation.tr("Lock")
                        onClicked: {
                            Quickshell.execDetached(["toren-shell", "lock"]);
                            sessionRoot.hide();
                        }
                        onFocusChanged: {
                            if (focus)
                                sessionRoot.subtitle = buttonText;
                        }
                        KeyNavigation.right: sessionLogout
                    }
                    SessionActionButton {
                        id: sessionLogout
                        buttonIcon: "logout"
                        buttonText: Translation.tr("Logout")
                        onClicked: {
                            ShellState.askSessionAction("logout", root.focusedScreen?.name ?? "");
                            sessionRoot.hide();
                        }
                        onFocusChanged: {
                            if (focus)
                                sessionRoot.subtitle = buttonText;
                        }
                        KeyNavigation.left: sessionLock
                        KeyNavigation.right: sessionHibernate.visible ? sessionHibernate : sessionReboot
                    }
                    SessionActionButton {
                        id: sessionHibernate
                        visible: Session.canHibernate
                        buttonIcon: "downloading"
                        buttonText: Translation.tr("Hibernate")
                        onClicked: {
                            Session.hibernate();
                            sessionRoot.hide();
                        }
                        onFocusChanged: {
                            if (focus)
                                sessionRoot.subtitle = buttonText;
                        }
                        KeyNavigation.left: sessionLogout
                        KeyNavigation.right: sessionReboot
                    }
                    SessionActionButton {
                        id: sessionReboot
                        buttonIcon: "restart_alt"
                        buttonText: Translation.tr("Reboot")
                        onClicked: {
                            ShellState.askSessionAction("reboot", root.focusedScreen?.name ?? "");
                            sessionRoot.hide();
                        }
                        onFocusChanged: {
                            if (focus)
                                sessionRoot.subtitle = buttonText;
                        }
                        KeyNavigation.left: sessionHibernate.visible ? sessionHibernate : sessionLogout
                        KeyNavigation.right: sessionShutdown
                    }
                    SessionActionButton {
                        id: sessionShutdown
                        buttonIcon: "power_settings_new"
                        buttonText: Translation.tr("Shutdown")
                        onClicked: {
                            ShellState.askSessionAction("shutdown", root.focusedScreen?.name ?? "");
                            sessionRoot.hide();
                        }
                        onFocusChanged: {
                            if (focus)
                                sessionRoot.subtitle = buttonText;
                        }
                        KeyNavigation.left: sessionReboot
                    }
                }

                DescriptionLabel {
                    Layout.alignment: Qt.AlignHCenter
                    text: sessionRoot.subtitle
                }
            }

            ColumnLayout {
                anchors {
                    top: contentColumn.bottom
                    topMargin: 10
                    horizontalCenter: contentColumn.horizontalCenter
                }
                spacing: 10

                Loader {
                    Layout.alignment: Qt.AlignHCenter
                    active: SessionWarnings.downloadRunning
                    visible: active
                    sourceComponent: DescriptionLabel {
                        text: Translation.tr("There might be a download in progress. Check your Downloads folder.")
                        textColor: Appearance.m3colors.m3onErrorContainer
                        color: Appearance.m3colors.m3errorContainer
                    }
                }

                Loader {
                    Layout.alignment: Qt.AlignHCenter
                    active: SessionWarnings.packageManagerRunning
                    visible: active
                    sourceComponent: DescriptionLabel {
                        text: Translation.tr("Your package manager is running")
                        textColor: Appearance.m3colors.m3onErrorContainer
                        color: Appearance.m3colors.m3errorContainer
                    }
                }
            }
        }
    }

    component DescriptionLabel: Rectangle {
        id: descriptionLabel
        property string text
        property color textColor: Appearance.colors.colOnTooltip
        color: Appearance.colors.colTooltip
        clip: true
        radius: Appearance.rounding.normal
        implicitHeight: descriptionLabelText.implicitHeight + 10 * 2
        implicitWidth: descriptionLabelText.implicitWidth + 15 * 2

        Behavior on implicitWidth {
            animation: Appearance.animation.elementMove.numberAnimation.createObject(this)
        }

        StyledText {
            id: descriptionLabelText
            anchors.centerIn: parent
            color: descriptionLabel.textColor
            text: descriptionLabel.text
        }
    }
}
