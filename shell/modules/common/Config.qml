import QtQuick
import Quickshell
import qs as Root

// The vendored components' settings, read from Asura's settings store.
//
// This is an overlay. Upstream's version of this file reads a JSON file of its
// own — it creates `~/.config/illogical-impulse/config.json` on first use and
// writes back into it whenever a component changes an option. Asura already
// has a settings store, GSettings, and exactly one writer, `asura config
// set`. Two stores would mean two answers to "what is the bar's height", and
// the one that wins would depend on which surface you asked.
//
// So this file exposes upstream's shape — `Config.options.bar.height`,
// `Config.options.appearance.fonts.main` — and every value comes from the
// wrapper at the shell root, which the shell's own reader fills from
// `asura config values`. It has no writer, because there is nothing here
// that owns a setting: a vendored component that wants to change one asks the
// desktop to, and the desktop writes GSettings.
//
// What is here is what a component would read to draw. Upstream's file has
// three hundred and seventy-nine options, most of them for surfaces Asura
// does not adopt — an AI assistant's model, a booru viewer's tags, a
// cheatsheet's keybindings. Those are not missing by oversight: an option for
// a surface that does not exist is a setting nobody can act on. When one of
// those surfaces is adopted, its options are added here and bound to the
// settings the desktop keeps.
pragma Singleton
pragma ComponentBehavior: Bound

Singleton {
    id: root

    property QtObject options: QtObject {
        property QtObject bar: QtObject {
            // The bar's own height, because the panel reserves exactly this
            // much space and a bar whose settings and geometry disagree is a
            // bar with a gap under it.
            property int height: Root.Config.barHeight
            property bool bottom: Root.Config.barPosition === Root.Config.edgeBottom
            property bool vertical: Root.Config.barPosition === Root.Config.edgeLeft
                || Root.Config.barPosition === Root.Config.edgeRight
            // Upstream calls an edged bar "attached" (0) and a floating one
            // "corned" (1). Asura's bar is attached: it is the top edge of
            // the screen, not a pill hovering above it.
            property int cornerStyle: 0
            // A verbosity switch that upstream's bar uses to decide whether
            // its centre module shows a clock, a date, or both. Asura's bar
            // shows both, so it is verbose.
            property bool verbose: true
            property bool showBackground: true
            // Every screen. An empty list means "all of them" upstream, and
            // Asura has no per-screen bar.
            property var screenList: []
            property QtObject autoHide: QtObject {
                property bool enable: false
                property QtObject showWhenPressingSuper: QtObject {
                    property bool enable: false
                    property int delay: 100
                }
            }
            property QtObject workspaces: QtObject {
                property bool showAppIcons: true
                property bool alwaysShowNumbers: false
            }
        }

        property QtObject dock: QtObject {
            // The dock's pinned applications, which are the desktop's
            // favourites because a dock is what a favourite is.
            property var pinnedApps: Root.Config.favorites
            property bool enabled: true
            property int iconSize: 48
        }

        property QtObject appearance: QtObject {
            property QtObject fonts: QtObject {
                property string main: Root.Theme.fontFamily
                property string numbers: Root.Theme.fontMonoFamily
                property string title: Root.Theme.fontFamily
                property string monospace: Root.Theme.fontMonoFamily
                property string reading: Root.Theme.fontFamily
                property string expressive: Root.Theme.fontFamily
                property string iconNerd: Root.Theme.fontFamily
            }
            property bool extraBackgroundTint: false
            // The screen corners. Upstream's three modes are 0 never,
            // 1 always, 2 except over a fullscreen window. Asura's setting is
            // on or off, and "on" is mode 2: a game or a video keeps its
            // whole frame.
            property int fakeScreenRounding: Root.Config.screenCorners ? 2 : 0
            // Panels are opaque. Upstream can composite them over the
            // wallpaper; Asura draws a surface with a hairline, so there is
            // no transparency to configure.
            property QtObject transparency: QtObject {
                property bool enable: false
                property bool automatic: false
                property real backgroundTransparency: 0
                property real contentTransparency: 0
            }
        }

        property QtObject background: QtObject {
            property string wallpaperPath: Root.Config.wallpaper
        }

        // The region selector: the OCR surface. The seven live knobs are the
        // schema's keys, read through the wrapper; the rest are constants,
        // each for the reason beside it.
        property QtObject regionSelector: QtObject {
            property QtObject targetRegions: QtObject {
                // What the selection snaps to while drawing. Windows on,
                // because that is the common case; layer surfaces opt in.
                property bool windows: Root.Config.regionSelectorTargetWindows
                property bool layers: Root.Config.regionSelectorTargetLayers
                property bool showLabel: Root.Config.regionSelectorTargetLabels
                // Content regions come from a text detection that upstream
                // runs through a Python virtualenv the vendored kit does not
                // carry, so there is nothing here to switch on.
                property bool content: false
                // How visible a snap target is while drawing. Constants: the
                // Hub has no control for a real number, and a setting the
                // Hub cannot show is a setting a user cannot change.
                property real opacity: 0.3
                property real contentRegionOpacity: 0.8
                // Padding past a snap target's edge, so the stroke drawn on
                // it is not half cut off by the edge itself.
                property int selectionPadding: Root.Config.regionSelectorSelectionPadding
            }
            property QtObject rect: QtObject {
                property bool showAimLines: Root.Config.regionSelectorAimLines
            }
            property QtObject circle: QtObject {
                property int strokeWidth: Root.Config.regionSelectorCircleStrokeWidth
                property int padding: Root.Config.regionSelectorCirclePadding
            }
            property QtObject annotation: QtObject {
                // Upstream annotates the crop by piping it into swappy or
                // satty. Asura's selector reaches the text recognition only,
                // and the desktop's annotate-after-capture belongs to the
                // screenshot tool, so the branch is never taken here.
                property bool useSatty: false
            }
        }

        // Upstream's image search uploads the crop somewhere and opens it in
        // an engine's lens. Asura's selector reaches the text recognition
        // only, so the URL is never read — it is here because the command
        // builder the recognition shares is whole, not half a file.
        property QtObject search: QtObject {
            property QtObject imageSearch: QtObject {
                property string imageSearchEngineBaseUrl: ""
            }
        }

        // Where upstream's copy action additionally saves the crop. Empty is
        // upstream's clipboard-only behaviour, which is also what Asura's
        // selector does: its one action reads the temporary screenshot in
        // place and never touches a save directory.
        property QtObject screenSnip: QtObject {
            property string savePath: ""
        }

        // The kit's interface language, read by its Translation service — the
        // selector's cursor guide phrases its hints through it. "auto"
        // follows the locale.
        property QtObject language: QtObject {
            property string ui: "auto"
        }

        // Scrolling. These are read by the kit's own flickable, so they exist
        // because a component is in use rather than because a surface might be
        // adopted one day. They are constants: the kit multiplies the wheel
        // delta by them to make a mouse wheel feel the same as a touchpad, and
        // the numbers are its own.
        property QtObject interactions: QtObject {
            // Upstream's workaround for panels that light a dead pixel row at
            // the screen's right and bottom edge: it pushes the corners one
            // pixel off-screen. Asura's hardware profiles have no such panel.
            property QtObject deadPixelWorkaround: QtObject {
                property bool enable: false
            }
            property QtObject scrolling: QtObject {
                property real touchpadScrollFactor: 100
                property real mouseScrollFactor: 50
                property real mouseScrollDeltaThreshold: 120
                property bool fasterTouchpadScroll: false
            }
        }

        // The screen corners can also open upstream's left and right sidebars
        // by click, hover or scroll. Asura has no left sidebar and opens the
        // right one from the bar and Super+N, so a corner is only a corner:
        // the interaction is off, and the rest are upstream's values so the
        // module reads defined numbers if it is ever turned on.
        property QtObject sidebar: QtObject {
            property QtObject cornerOpen: QtObject {
                property bool enable: false
                property bool bottom: false
                property bool valueScroll: true
                property bool clickless: false
                property bool clicklessCornerEnd: true
                property int clicklessCornerVerticalOffset: 1
                property int cornerRegionWidth: 250
                property int cornerRegionHeight: 5
                property bool visualize: false
            }
        }

        // Upstream's components read this to know whether their settings file
        // had loaded. Asura's settings are a GSettings read that is either
        // there or is not, and the defaults are the schema's, so there is no
        // half-loaded state to report.
        property bool ready: true
    }
}
