import QtQuick
import Quickshell
import Quickshell.Io
import qs.Ui
import qs.Ui.Singletons
import ".."
import "../schema/DesktopShellPage.js" as DesktopShellSchema

// Desktop: the default desktop's own settings (bar, dock, colours, clock,
// notifications, lock screen, battery, night light), read from and written to
// the file that desktop reads, ~/.config/illogical-impulse/config.json. The
// desktop watches that file and applies a change at once, so an edit here is
// written straight away rather than held for the action bar.
//
// Only the keys this page shows are touched: the file is read whole, the one
// value set at its dotted path, and the whole file written back, so the rest
// of the desktop's settings (and anything the desktop's own Settings app
// wrote) stay as they were.
Item {
    id: pg
    property var hub

    readonly property string pTitle: I18n.tr("Desktop")
    readonly property string pEyebrow: I18n.tr("DESKTOP")
    readonly property string pBlurb: I18n.tr("The bar, the dock, colours, the clock, notifications, the lock screen, battery warnings and night light. Changes apply at once.")
    function focusKey(k) { sp.focusKey(k) }

    readonly property string configHome: Quickshell.env("XDG_CONFIG_HOME") || (Quickshell.env("HOME") + "/.config")
    readonly property string path: pg.configHome + "/illogical-impulse/config.json"

    // Rounded app corners: kept outside config.json, in the file the GTK
    // stylesheets import (~/.config/gtk-N.0/toren-corners.css). It imports
    // Toren's square-corner rules, or holds a comment when the user wants the
    // apps' own rounded corners. toren-session writes the square one at login
    // when the file is missing, so square is the default.
    property bool gtkRounded: false
    function cornersCss(v, rounded) {
        return rounded
            ? "/* Toren: apps keep their own rounded corners (Settings -> Desktop -> Look). */\n"
            : "@import url(\"file:///usr/share/toren/apps/gtk/sharp-gtk" + v + ".css\");\n";
    }
    function setGtkRounded(rounded) {
        pg.gtkRounded = rounded;
        gtk3Corners.setText(pg.cornersCss(3, rounded));
        gtk4Corners.setText(pg.cornersCss(4, rounded));
        var view = Object.assign({}, pg.draft);
        view["toren.gtkRoundedCorners"] = rounded;
        pg.draft = view;
    }

    // the file as last read, and the page's view of it (shown values, keyed by
    // the schema's dotted keys)
    property var doc: ({})
    property var draft: ({})
    readonly property var defaults: {
        var d = {};
        for (var i = 0; i < DesktopShellSchema.rows.length; i++) {
            var r = DesktopShellSchema.rows[i];
            d[r.key] = pg.shown(r, r.def);
        }
        return d;
    }

    function rowFor(key) {
        for (var i = 0; i < DesktopShellSchema.rows.length; i++)
            if (DesktopShellSchema.rows[i].key === key) return DesktopShellSchema.rows[i];
        return null;
    }
    function getPath(o, key) {
        var parts = key.split(".");
        for (var i = 0; i < parts.length; i++) {
            if (o === null || o === undefined || typeof o !== "object") return undefined;
            o = o[parts[i]];
        }
        return o;
    }
    function setPath(o, key, v) {
        var parts = key.split(".");
        for (var i = 0; i < parts.length - 1; i++) {
            if (o[parts[i]] === null || typeof o[parts[i]] !== "object") o[parts[i]] = {};
            o = o[parts[i]];
        }
        o[parts[parts.length - 1]] = v;
    }
    // a stored value as the control shows it (a `map` row shows its label)
    function shown(r, stored) {
        if (!r.map) return stored;
        for (var label in r.map)
            if (r.map[label] === stored) return label;
        return stored;
    }
    // a shown value as the file stores it
    function stored(r, v) {
        if (r.map && r.map[v] !== undefined) return r.map[v];
        return v;
    }

    function adopt(text) {
        var d = {};
        try { d = JSON.parse(text); } catch (e) { d = {}; }
        if (d === null || typeof d !== "object") d = {};
        pg.doc = d;
        var view = {};
        for (var i = 0; i < DesktopShellSchema.rows.length; i++) {
            var r = DesktopShellSchema.rows[i];
            if (r.store === "gtk-corners") { view[r.key] = pg.gtkRounded; continue; }
            if (r.store === "action") continue;
            if (r.store === "charge-limit") { view[r.key] = pg.chargeLimit; continue; }
            var v = pg.getPath(d, r.key);
            view[r.key] = pg.shown(r, v === undefined ? r.def : v);
        }
        pg.draft = view;
    }

    function edit(key, v) {
        var r = pg.rowFor(key);
        if (!r) return;
        if (r.store === "gtk-corners") { pg.setGtkRounded(v === true); return; }
        if (r.store === "charge-limit") {
            // root writes it (toren power threshold); pkexec asks for the password
            setCharge.command = ["pkexec", "toren", "power", "threshold", String(v)];
            setCharge.running = true;
            return;
        }
        var d = JSON.parse(JSON.stringify(pg.doc));
        pg.setPath(d, key, pg.stored(r, v));
        pg.doc = d;
        var view = Object.assign({}, pg.draft);
        view[key] = v;
        pg.draft = view;
        cfg.setText(JSON.stringify(d, null, 2) + "\n");
    }

    FileView {
        id: cfg
        path: pg.path
        blockLoading: true
        watchChanges: true
        printErrors: false
        atomicWrites: true
        onFileChanged: reload()
        onLoaded: pg.adopt(text())
        onLoadFailed: pg.adopt("{}")
    }

    FileView {
        id: gtk4Corners
        path: pg.configHome + "/gtk-4.0/toren-corners.css"
        blockLoading: true
        watchChanges: true
        printErrors: false
        atomicWrites: true
        onFileChanged: reload()
        onLoaded: {
            pg.gtkRounded = text().indexOf("sharp-gtk") < 0;
            var view = Object.assign({}, pg.draft);
            view["toren.gtkRoundedCorners"] = pg.gtkRounded;
            pg.draft = view;
        }
    }
    // The battery charge limit the user chose (toren power threshold), or 80,
    // Toren's default, when they have not.
    property int chargeLimit: 80
    FileView {
        path: "/etc/toren/charge-limit"
        watchChanges: true
        printErrors: false
        onFileChanged: reload()
        onLoaded: {
            var n = parseInt(text().trim());
            pg.chargeLimit = isNaN(n) ? 80 : n;
            var view = Object.assign({}, pg.draft);
            view["toren.chargeLimit"] = pg.chargeLimit;
            pg.draft = view;
        }
    }
    Process { id: setCharge }

    FileView {
        id: gtk3Corners
        path: pg.configHome + "/gtk-3.0/toren-corners.css"
        printErrors: false
        atomicWrites: true
    }

    SchemaPage {
        id: sp
        anchors.fill: parent
        schema: DesktopShellSchema.rows
        draft: pg.draft
        defaults: pg.defaults
        advanced: pg.hub ? pg.hub.advanced : true
        title: pg.pTitle
        eyebrow: pg.pEyebrow
        blurb: pg.pBlurb
        query: pg.hub ? pg.hub.query : ""
        onEdited: (k, v) => pg.edit(k, v)
    }
}
