import QtQuick
import "Singletons"

// boolean. two named modes are a Seg, not a switch.
Rectangle {
    id: sw
    property bool on: false
    signal toggled(bool v)

    activeFocusOnTab: true
    function activate() { toggled(!on) }
    Keys.onPressed: event => {
        if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter || event.key === Qt.Key_Space) {
            activate();
            event.accepted = true;
        }
    }

    implicitWidth: Tokens.cards ? 44 : 54
    implicitHeight: 24
    radius: Tokens.cards ? height / 2 : Tokens.radius
    // off sits on a faint tint with a dim knob, on brightens the track and fills
    // the knob: two signals for one state, so it reads at a glance. The card
    // look is a round switch whose track fills with the accent when on.
    color: Tokens.cards ? (sw.on ? Tokens.sun : (tap.pressed ? Tokens.tint16 : Tokens.tint10))
        : (tap.pressed ? Tokens.tint16 : (sw.on ? Tokens.tint16 : Tokens.tint5))
    border.width: Tokens.cards ? 0 : Tokens.border
    border.color: activeFocus ? Tokens.bone : (hh.hovered ? Tokens.lineStrong : Tokens.line)
    antialiasing: Tokens.cards
    Behavior on border.color { ColorAnimation { duration: Tokens.snap } }
    Behavior on color { ColorAnimation { duration: Tokens.snap } }

    Rectangle {
        width: Tokens.cards ? 18 : 25
        height: Tokens.cards ? 18 : 17
        y: Tokens.cards ? 3 : 3
        x: sw.on ? parent.width - width - 3 : 3
        radius: Tokens.cards ? height / 2 : Tokens.radius
        antialiasing: Tokens.cards
        // off is a dim knob on a hairline track, not an empty outline: the state
        // has to be readable at a glance, which an outlined box is not.
        color: Tokens.cards ? (sw.on ? Tokens.onSun : Tokens.inkDim) : (sw.on ? Tokens.ink : Tokens.inkDim)
        border.width: 0
        Behavior on x { NumberAnimation { duration: Tokens.snap; easing.type: Tokens.easeSnap } }
    }
    HoverHandler { id: hh; cursorShape: Qt.PointingHandCursor }
    TapHandler { id: tap; onTapped: sw.toggled(!sw.on) }
}
