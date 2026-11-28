import qs.services
import qs.modules.common
import qs.modules.common.widgets
import QtQuick
import QtQuick.Layouts

Item {
    id: root

    implicitHeight: contentColumn.implicitHeight
    implicitWidth: contentColumn.implicitWidth

    function formatTime(totalSeconds) {
        const s = Math.max(0, Math.floor(totalSeconds));
        const hrs = Math.floor(s / 3600);
        const mins = Math.floor((s % 3600) / 60).toString().padStart(2, '0');
        const secs = Math.floor(s % 60).toString().padStart(2, '0');
        return hrs > 0 ? `${hrs}:${mins}:${secs}` : `${mins}:${secs}`;
    }

    ColumnLayout {
        id: contentColumn
        anchors.fill: parent
        spacing: 8

        CircularProgress {
            Layout.alignment: Qt.AlignHCenter
            lineWidth: 7
            value: TimerService.countdownDuration > 0 ? (TimerService.countdownSecondsLeft / TimerService.countdownDuration) : 0
            implicitSize: 158
            enableAnimation: true

            ColumnLayout {
                anchors.centerIn: parent
                spacing: 0

                StyledText {
                    Layout.alignment: Qt.AlignHCenter
                    text: root.formatTime(TimerService.countdownSecondsLeft)
                    font.pixelSize: 34
                    color: Appearance.m3colors.m3onSurface
                }
                StyledText {
                    Layout.alignment: Qt.AlignHCenter
                    text: TimerService.countdownRunning ? Translation.tr("Running")
                        : (TimerService.countdownSecondsLeft === 0 ? Translation.tr("Done") : Translation.tr("Timer"))
                    font.pixelSize: Appearance.font.pixelSize.small
                    color: Appearance.colors.colSubtext
                }
            }
        }

        // Quick preset chips + +/-1m adjust
        RowLayout {
            Layout.alignment: Qt.AlignHCenter
            spacing: 4
            enabled: !TimerService.countdownRunning
            opacity: enabled ? 1.0 : 0.45

            Repeater {
                model: [
                    { label: "-1m", delta: -60 },
                    { label: "1m",  secs: 60 },
                    { label: "5m",  secs: 300 },
                    { label: "10m", secs: 600 },
                    { label: "15m", secs: 900 },
                    { label: "+1m", delta: 60 }
                ]
                delegate: RippleButton {
                    required property var modelData
                    readonly property bool isPreset: modelData.secs !== undefined
                    readonly property bool selected: isPreset && TimerService.countdownDuration === modelData.secs
                    implicitWidth: 44
                    implicitHeight: 26
                    buttonRadius: Appearance.rounding.small
                    colBackground: selected ? Appearance.colors.colSecondaryContainer : Appearance.colors.colLayer2
                    colBackgroundHover: selected ? Appearance.colors.colSecondaryContainerHover : Appearance.colors.colLayer2Hover
                    onClicked: {
                        if (isPreset) {
                            TimerService.setCountdownDuration(modelData.secs);
                        } else {
                            TimerService.setCountdownDuration(TimerService.countdownDuration + modelData.delta);
                        }
                    }

                    contentItem: StyledText {
                        anchors.centerIn: parent
                        text: modelData.label
                        font.pixelSize: Appearance.font.pixelSize.smaller
                        color: selected ? Appearance.colors.colOnSecondaryContainer : Appearance.colors.colOnLayer2
                    }
                }
            }
        }

        // Start/Pause & Reset buttons
        RowLayout {
            Layout.alignment: Qt.AlignHCenter
            spacing: 10

            RippleButton {
                implicitHeight: 34
                implicitWidth: 90
                onClicked: TimerService.toggleCountdown()
                colBackground: TimerService.countdownRunning ? Appearance.colors.colSecondaryContainer : Appearance.colors.colPrimary
                colBackgroundHover: TimerService.countdownRunning ? Appearance.colors.colSecondaryContainerHover : Appearance.colors.colPrimaryHover

                contentItem: StyledText {
                    anchors.centerIn: parent
                    horizontalAlignment: Text.AlignHCenter
                    text: TimerService.countdownRunning
                        ? Translation.tr("Pause")
                        : (TimerService.countdownSecondsLeft === TimerService.countdownDuration || TimerService.countdownSecondsLeft === 0
                            ? Translation.tr("Start") : Translation.tr("Resume"))
                    color: TimerService.countdownRunning ? Appearance.colors.colOnSecondaryContainer : Appearance.colors.colOnPrimary
                }
            }

            RippleButton {
                implicitHeight: 34
                implicitWidth: 90
                onClicked: TimerService.resetCountdown()
                enabled: TimerService.countdownRunning || TimerService.countdownSecondsLeft !== TimerService.countdownDuration
                colBackground: Appearance.colors.colErrorContainer
                colBackgroundHover: Appearance.colors.colErrorContainerHover
                colRipple: Appearance.colors.colErrorContainerActive

                contentItem: StyledText {
                    anchors.centerIn: parent
                    horizontalAlignment: Text.AlignHCenter
                    text: Translation.tr("Reset")
                    color: Appearance.colors.colOnErrorContainer
                }
            }
        }
    }
}
