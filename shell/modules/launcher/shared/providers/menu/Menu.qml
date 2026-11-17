import QtQuick
import Quickshell
import Quickshell.Io
import qs.Ui.Singletons
import "../../Singletons"
import "../../../../../utils/fuzzy.js" as Fuzzy
import "catalog.js" as Catalog
import ".."

// ASURA Menu provider: single "do everything" menu tree covering Apps,
// Learn, Style, Setup, Install/Remove, Update, Capture, and System.
// Every leaf calls an existing tool or is visibly marked disabled when
// its binary is absent.
Provider {
    id: menu

    providerId: "menu"
    prefix: ":menu"
    defaultProvider: false

    property string activeCategory: "All"
    onActiveCategoryChanged: Dispatcher.notifyAsync()

    // The binaries found by PATH probe
    property var present: ({})
    property bool probed: false
    Process {
        running: true
        command: ["sh", "-c", 'for c; do command -v "$c" >/dev/null 2>&1 && echo "$c"; done', "sh"].concat(Catalog.commands())
        stdout: SplitParser {
            onRead: name => menu.present[name] = true
        }
        onExited: {
            menu.probed = true;
            Dispatcher.notifyAsync();
        }
    }

    function isAvailable(entry) {
        if (entry.mode)
            return true;
        if (!entry.exec || entry.exec.length === 0)
            return false;
        var binary = entry.exec[0];
        if (binary === "sh" || binary === "sudo")
            return true;
        if (menu.probed && menu.present[binary] !== true)
            return false;
        return true;
    }

    function rowFor(entry) {
        var available = isAvailable(entry);
        var sub = I18n.tr(entry.desc);
        if (!available) {
            sub = sub + " (" + I18n.tr("Not available") + ")";
        }

        return {
            id: "menu:" + entry.id,
            title: I18n.tr(entry.name),
            subtitle: sub,
            icon: entry.icon || "system-run",
            type: entry.category,
            category: entry.category,
            disabled: !available,
            score: 0,
            actions: [{
                id: "run",
                name: available ? I18n.tr("Open") : I18n.tr("Not available"),
                icon: "",
                mode: entry.mode || "",
                closeOnExecute: !entry.mode,
                execute: function () {
                    if (!available)
                        return;
                    if (entry.exec) {
                        Spawn.run(entry.exec);
                    }
                }
            }]
        };
    }

    function query(text) {
        var pool = Catalog.CATALOG.filter(function (entry) {
            return menu.activeCategory === "All" || entry.category === menu.activeCategory;
        });
        var q = (text || "").trim().toLowerCase();
        var rows = [];
        for (var i = 0; i < pool.length; i++) {
            var item = pool[i];
            if (q.length === 0) {
                rows.push(rowFor(item));
            } else {
                var score = Fuzzy.score({
                    name: item.name,
                    keywords: [item.category, item.desc]
                }, q);
                if (score < 99) {
                    var r = rowFor(item);
                    r.score = score;
                    rows.push(r);
                }
            }
        }
        return rows;
    }

    Component.onCompleted: Dispatcher.register(menu);
}
