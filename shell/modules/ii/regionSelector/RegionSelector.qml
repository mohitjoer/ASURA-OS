pragma ComponentBehavior: Bound
import qs
import qs.modules.common
import QtQuick
import Quickshell

// The region selector, as Toren ships it: draw a rectangle or a circle on the
// screen, and the text inside it is recognised and copied to the clipboard.
//
// This is an overlay. Upstream's version of this file is the entry point for
// five actions — copy, annotate, image search, text recognition and screen
// recording — and registers an IPC handler and a global shortcut for each.
// Toren's region selector reaches exactly one of them, the text recognition,
// because the desktop's other halves of those actions already exist: the
// screenshot tool owns capture and annotation, the colour picker owns picking,
// and a region recording is not something the desktop ships. An entry point
// for an action the desktop does not have is a control that runs a command
// that is not there.
//
// The drawing itself is upstream's (RegionSelection, in this directory,
// unmodified): the rectangle or circle, the window snapping while you draw,
// and the Esc to dismiss. What this file adds is the flag that opens it and
// the one action it carries.
//
// The keybind reaches this through the shell's own IPC bridge, as
// `toren ipc call toren ocr` — the same bridge every shell surface is called
// through — so there is no second handler to keep a target name honest with.
Scope {
    id: root

    // Recognition is the only action this surface reaches, so it is not a
    // property the callers set: it is what the surface is.
    readonly property var action: RegionSelection.SnipAction.CharRecognition
    property var selectionMode: RegionSelection.SelectionMode.RectCorners

    function dismiss() {
        GlobalStates.regionSelectorOpen = false
    }

    function ocr() {
        root.selectionMode = RegionSelection.SelectionMode.RectCorners
        GlobalStates.regionSelectorOpen = true
    }

    Variants {
        model: Quickshell.screens
        delegate: Loader {
            id: regionSelectorLoader
            required property var modelData
            active: GlobalStates.regionSelectorOpen

            sourceComponent: RegionSelection {
                screen: regionSelectorLoader.modelData
                onDismiss: root.dismiss()
                action: root.action
                selectionMode: root.selectionMode
            }
        }
    }
}
