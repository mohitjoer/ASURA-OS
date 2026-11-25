pragma Singleton
pragma ComponentBehavior: Bound

// Which of the shell's surfaces are open.
//
// Upstream's shell keeps this state in a singleton at the shell root, and
// the vendored components read it: the bar asks whether the overview is open
// before it draws itself over it, the dock asks whether the screen is
// locked before it hides.
//
// Upstream's own file is not vendored, for two reasons. It is upstream's
// shell root, and the shell root is Toren's — Toren's shell is a
// different shell that happens to draw some of the same surfaces. And it
// imports the whole service graph, which closes a cycle through the modules
// directory: the services reach back into the modules, and the modules
// reach back into the root.
//
// So this is Toren's, at upstream's path, carrying the same flags.
//
// A flag is a decision, not a reading: something opens a surface and
// something closes it, and a flag that merely reported a state would have
// two writers and drift. The shell's own surfaces set these.

import Quickshell
import QtQuick

Singleton {
    id: root

    // ----------------------------------------------------------------- panels
    //
    // One flag per surface that can be open at the same time as another.
    // A surface that cannot be open twice does not need a flag.
    property bool barOpen: false
    property bool calendarOpen: false
    property bool overviewOpen: false
    property bool sidebarLeftOpen: false
    property bool sidebarRightOpen: false
    property bool mediaControlsOpen: false
    property bool osdVolumeOpen: false
    property bool osdBrightnessOpen: false
    property bool oskOpen: false
    property bool sessionOpen: false

    // The session screen is the one flag in this list whose surface is a
    // full-screen overlay layer over the focused screen rather than a panel
    // beside them. The sidebar's power button opens it and the
    // Ctrl+Alt+Delete bind toggles it; it closes itself, on a click
    // anywhere, Escape, or the lock engaging.

    // The region selector. It covers every screen while it is open, so it is
    // not one of the panel flags above: there is no other surface it shares
    // a screen with while it draws, and the selection ends it — the surface
    // itself sets this false on dismiss.
    property bool regionSelectorOpen: false

    // ------------------------------------------------------------------ lock
    //
    // Whether the session is locked. A locked screen hides the surfaces that
    // would let a person act on the desktop behind it.
    property bool screenLocked: false

    // Whether the last unlock attempt failed.
    property bool screenUnlockFailed: false

    // Whether the lock screen is showing characters, which is how a person
    // is told that what they typed reached the lock screen at all.
    property bool screenLockContainsCharacters: false

    // --------------------------------------------------------------- super key
    //
    // The Super key is held rather than pressed, so a chord that ends with
    // it has to wait for the release before it acts. These say whether the
    // key is down and whether a release might trigger something.
    property bool superDown: false
    property bool superReleaseMightTrigger: false

    // ------------------------------------------------------------------ focus
    //
    // Whether a surface has taken the keyboard. A surface that has is the
    // one that gets the next keypress.
    property bool focusGrabActive: false
}
