pragma Singleton
pragma ComponentBehavior: Bound

import qs.services
import qs.modules.common

import Quickshell
import Quickshell.Io
import QtQuick

/**
 * Timer and stopwatch manager.
 */
Singleton {
    id: root

    property bool stopwatchRunning: Persistent.states.timer.stopwatch.running
    property int stopwatchTime: 0
    property int stopwatchStart: Persistent.states.timer.stopwatch.start
    property var stopwatchLaps: Persistent.states.timer.stopwatch.laps

    property bool countdownRunning: false
    property int countdownDuration: 300
    property int countdownSecondsLeft: 300
    property int countdownEndTime: 0

    Component.onCompleted: {
        if (!stopwatchRunning)
            stopwatchReset();
    }

    function getCurrentTimeInSeconds() {
        return Math.floor(Date.now() / 1000);
    }

    function getCurrentTimeIn10ms() {
        return Math.floor(Date.now() / 10);
    }

    // Stopwatch
    function refreshStopwatch() {  // Stopwatch stores time in 10ms
        stopwatchTime = getCurrentTimeIn10ms() - stopwatchStart;
    }

    Timer {
        id: stopwatchTimer
        interval: 10
        running: root.stopwatchRunning
        repeat: true
        onTriggered: refreshStopwatch()
    }

    function toggleStopwatch() {
        if (root.stopwatchRunning)
            stopwatchPause();
        else
            stopwatchResume();
    }

    function stopwatchPause() {
        Persistent.states.timer.stopwatch.running = false;
    }

    function stopwatchResume() {
        if (stopwatchTime === 0) Persistent.states.timer.stopwatch.laps = [];
        Persistent.states.timer.stopwatch.running = true;
        Persistent.states.timer.stopwatch.start = getCurrentTimeIn10ms() - stopwatchTime;
    }

    function stopwatchReset() {
        stopwatchTime = 0;
        Persistent.states.timer.stopwatch.laps = [];
        Persistent.states.timer.stopwatch.running = false;
    }

    function stopwatchRecordLap() {
        Persistent.states.timer.stopwatch.laps.push(stopwatchTime);
    }

    // Countdown Timer
    function refreshCountdown() {
        const left = countdownEndTime - getCurrentTimeInSeconds();
        if (left <= 0) {
            countdownRunning = false;
            countdownSecondsLeft = 0;
            Quickshell.execDetached(["notify-send", "Timer", Translation.tr("Time's up!"), "-a", "Shell"]);
            if (Config.options.sounds.pomodoro) {
                Audio.playSystemSound("alarm-clock-elapsed");
            }
            return;
        }
        countdownSecondsLeft = left;
    }

    Timer {
        id: countdownTimer
        interval: 200
        running: root.countdownRunning
        repeat: true
        onTriggered: refreshCountdown()
    }

    function setCountdownDuration(seconds) {
        if (countdownRunning)
            return;
        countdownDuration = Math.max(60, Math.min(36000, seconds));
        countdownSecondsLeft = countdownDuration;
    }

    function toggleCountdown() {
        if (countdownRunning) {
            refreshCountdown();
            countdownRunning = false;
        } else {
            if (countdownSecondsLeft <= 0) {
                countdownSecondsLeft = countdownDuration;
            }
            countdownEndTime = getCurrentTimeInSeconds() + countdownSecondsLeft;
            countdownRunning = true;
        }
    }

    function resetCountdown() {
        countdownRunning = false;
        countdownSecondsLeft = countdownDuration;
    }
}
