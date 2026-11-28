pragma ComponentBehavior: Bound
import QtQuick
import QtQuick.Controls
import Quickshell
import Quickshell.Io
import qs.Ui
import qs.Ui.Singletons

Rectangle {
    id: root

    color: Tokens.paper
    focus: true

    // ── wizard state ────────────────────────────────────────────────────────
    readonly property var steps: [
        { "key": "region",  "title": I18n.tr("Welcome & Region"), "desc": I18n.tr("Choose your language, keyboard layout, and timezone.") },
        { "key": "storage", "title": I18n.tr("Disk & Storage"),   "desc": I18n.tr("Select the target drive, filesystem, and optional encryption.") },
        { "key": "account", "title": I18n.tr("Your Account"),     "desc": I18n.tr("Create your user account and name this computer.") },
        { "key": "review",  "title": I18n.tr("Review & Install"), "desc": I18n.tr("Confirm your choices before writing to disk.") },
        { "key": "install", "title": I18n.tr("Installing"),       "desc": I18n.tr("Installing Toren Linux onto your drive.") }
    ]
    property int step: 0

    // Region
    readonly property var locales: [
        { "code": "en_US.UTF-8", "label": "English (US)" },
        { "code": "en_GB.UTF-8", "label": "English (UK)" },
        { "code": "de_DE.UTF-8", "label": "Deutsch" },
        { "code": "fr_FR.UTF-8", "label": "Fran\u00e7ais" },
        { "code": "es_ES.UTF-8", "label": "Espa\u00f1ol" },
        { "code": "ja_JP.UTF-8", "label": "\u65e5\u672c\u8a9e" }
    ]
    readonly property var keymaps: ["us", "uk", "de", "fr", "es", "jp"]
    readonly property var timezones: [
        "UTC", "America/New_York", "America/Los_Angeles",
        "Europe/London", "Europe/Berlin", "Asia/Kolkata", "Asia/Tokyo"
    ]
    property string locale: "en_US.UTF-8"
    property string keymap: "us"
    property string timezone: "UTC"

    // Storage
    property var disks: []
    property bool disksLoading: true
    property string diskError: ""
    property string selectedDisk: ""
    property string installMode: "wipe"        // "wipe" | "alongside"
    property string filesystem: "btrfs"        // "btrfs" | "ext4"
    property bool encrypt: false
    property string luksPass: ""
    property string luksConfirm: ""

    // Account
    property string username: "toren"
    property string hostname: "toren"
    property string password: ""
    property string passwordConfirm: ""

    // Install execution
    property bool installing: false
    property bool installDone: false
    property bool installOk: false
    property int installProgress: 0
    property string installStatus: I18n.tr("Preparing installation...")
    property string installLog: ""
    property string tempConfigPath: ""

    readonly property bool storageValid: {
        if (selectedDisk === "") return false;
        if (encrypt && (luksPass.length < 4 || luksPass !== luksConfirm)) return false;
        return true;
    }
    readonly property bool accountValid: {
        if (!/^[a-z_][a-z0-9_-]{0,31}$/.test(username)) return false;
        if (!/^[a-zA-Z0-9][a-zA-Z0-9-]{0,62}$/.test(hostname)) return false;
        if (password.length < 1 || password !== passwordConfirm) return false;
        return true;
    }
    readonly property bool canAdvance: {
        if (step === 0) return timezone.trim().length > 0;
        if (step === 1) return storageValid;
        if (step === 2) return accountValid;
        if (step === 3) return storageValid && accountValid;
        return false;
    }

    function formatBytes(b) {
        if (!b || b <= 0) return "0 GB";
        const gb = b / 1000000000;
        if (gb >= 1000) return (gb / 1000).toFixed(2) + " TB";
        return gb.toFixed(1) + " GB";
    }

    function refreshDisks() {
        disksLoading = true;
        diskError = "";
        diskProc.running = true;
    }

    function parseDisks(raw) {
        try {
            const parsed = JSON.parse(raw);
            const list = [];
            const devs = (parsed && parsed.blockdevices) ? parsed.blockdevices : [];
            for (let i = 0; i < devs.length; i++) {
                const d = devs[i];
                const name = String(d.name || "");
                const type = String(d.type || "");
                if (!name || type === "rom" || name.indexOf("loop") === 0 || name.indexOf("zram") === 0 || name.indexOf("sr") === 0)
                    continue;
                const path = String(d.path || ("/dev/" + name));
                const size = Number(d.size || 0);
                if (size < 4000000000) continue;
                list.push({
                    "name": name,
                    "path": path,
                    "size": size,
                    "sizeText": formatBytes(size),
                    "model": String(d.model || "Disk").trim(),
                    "tran": String(d.tran || "").toUpperCase()
                });
            }
            disks = list;
            if (list.length > 0) {
                let found = false;
                for (let i = 0; i < list.length; i++) {
                    if (list[i].path === selectedDisk) { found = true; break; }
                }
                if (!found) selectedDisk = list[0].path;
            } else {
                selectedDisk = "";
            }
        } catch (e) {
            disks = [];
            diskError = I18n.tr("Could not enumerate disks.");
        }
        disksLoading = false;
    }

    function appendLog(line) {
        const clean = line.replace(/\x1b\[[0-9;]*m/g, "");
        if (!clean) return;
        installLog += clean + "\n";
        const lower = clean.toLowerCase();
        if (lower.indexOf("partition") >= 0 || lower.indexOf("formatting") >= 0) {
            installProgress = Math.max(installProgress, 20);
            installStatus = I18n.tr("Partitioning and formatting drive...");
        } else if (lower.indexOf("copying") >= 0 || lower.indexOf("unsquash") >= 0 || lower.indexOf("system") >= 0) {
            installProgress = Math.max(installProgress, 45);
            installStatus = I18n.tr("Installing Toren Linux system files...");
        } else if (lower.indexOf("user") >= 0 || lower.indexOf("locale") >= 0 || lower.indexOf("timezone") >= 0) {
            installProgress = Math.max(installProgress, 70);
            installStatus = I18n.tr("Configuring user account and locale...");
        } else if (lower.indexOf("bootloader") >= 0 || lower.indexOf("grub") >= 0 || lower.indexOf("initramfs") >= 0) {
            installProgress = Math.max(installProgress, 88);
            installStatus = I18n.tr("Installing bootloader and initramfs...");
        } else if (lower.indexOf("verification") >= 0 || lower.indexOf("complete") >= 0) {
            installProgress = Math.max(installProgress, 96);
            installStatus = I18n.tr("Verifying installation...");
        }
    }

    function startInstall() {
        step = 4;
        installing = true;
        installDone = false;
        installOk = false;
        installProgress = 5;
        installStatus = I18n.tr("Starting installer...");
        installLog = "";

        const cfg = [
            "disk: " + selectedDisk,
            "mode: " + installMode,
            "filesystem: " + filesystem,
            "encryption: " + (encrypt ? "yes" : "no"),
            encrypt ? ("luks_password: " + luksPass) : "",
            "hostname: " + hostname.trim(),
            "user: " + username.trim(),
            "password: " + password,
            "timezone: " + timezone.trim(),
            "locale: " + locale,
            "keyboard: " + keymap
        ].filter(Boolean).join("\n") + "\n";

        const b64 = Qt.btoa(cfg);
        const runner = [
            'set -eu',
            'CFG=$(mktemp /tmp/toren-installer-XXXXXX.conf)',
            'chmod 600 "$CFG"',
            'trap "rm -f \\"$CFG\\"" EXIT',
            'printf %s "$1" | base64 -d > "$CFG"',
            'BIN=/usr/bin/toren-installer',
            '[ -x "$BIN" ] || BIN=/usr/share/toren/installation/toren-installer.sh',
            'if [ "$(id -u)" = 0 ]; then',
            '  exec "$BIN" --config "$CFG"',
            'elif sudo -n true 2>/dev/null; then',
            '  exec sudo "$BIN" --config "$CFG"',
            'else',
            '  exec pkexec "$BIN" --config "$CFG"',
            'fi'
        ].join("\n");

        installProc.command = ["sh", "-c", runner, "toren-installer-ui", b64];
        installProc.running = true;
    }

    Component.onCompleted: refreshDisks()

    Process {
        id: diskProc
        command: ["lsblk", "-J", "-d", "-b", "-o", "NAME,PATH,SIZE,MODEL,TRAN,RM,TYPE"]
        stdout: StdioCollector {
            onStreamFinished: root.parseDisks(this.text)
        }
        onExited: (code) => {
            if (code !== 0 && root.disksLoading) {
                root.disksLoading = false;
                root.diskError = I18n.tr("lsblk failed to list drives.");
            }
        }
    }

    Process {
        id: installProc
        stdout: SplitParser { onRead: line => root.appendLog(line) }
        stderr: SplitParser { onRead: line => root.appendLog(line) }
        onExited: (code) => {
            root.installing = false;
            root.installDone = true;
            root.installOk = (code === 0);
            if (code === 0) {
                root.installProgress = 100;
                root.installStatus = I18n.tr("Installation complete!");
            } else {
                root.installStatus = I18n.tr("Installation failed (exit code %1).").arg(code);
            }
        }
    }

    Process {
        id: rebootProc
        command: ["sh", "-c", "systemctl reboot || loginctl reboot || sudo reboot"]
    }

    // ── layout: left step rail + main step body ─────────────────────────────
    Row {
        anchors.fill: parent

        // Left rail
        Rectangle {
            width: 248
            height: parent.height
            color: Tokens.tint5

            Rectangle {
                anchors.right: parent.right
                width: Tokens.border
                height: parent.height
                color: Tokens.line
            }

            Column {
                anchors.fill: parent
                anchors.margins: Tokens.s6
                spacing: Tokens.s6

                // Brand lockup
                Row {
                    spacing: Tokens.s3
                    Image {
                        width: 24; height: 24
                        anchors.verticalCenter: parent.verticalCenter
                        source: Qt.resolvedUrl("../assets/toren-logo-on-dark.svg")
                        sourceSize.width: 48; sourceSize.height: 48
                        fillMode: Image.PreserveAspectFit
                        smooth: true
                    }
                    Column {
                        spacing: 2
                        Text {
                            text: "Toren Linux"
                            color: Tokens.ink
                            font.family: Tokens.display
                            font.pixelSize: 16
                            font.weight: Font.DemiBold
                        }
                        Text {
                            text: I18n.tr("System Installer")
                            color: Tokens.inkMuted
                            font.family: Tokens.ui
                            font.pixelSize: Tokens.fMicro
                        }
                    }
                }

                Rectangle { width: parent.width; height: Tokens.border; color: Tokens.line }

                // Step list
                Column {
                    width: parent.width
                    spacing: Tokens.s2

                    Repeater {
                        model: root.steps
                        delegate: Rectangle {
                            required property var modelData
                            required property int index
                            readonly property bool active: root.step === index
                            readonly property bool done: root.step > index || (index === 4 && root.installOk)
                            width: parent.width
                            height: 42
                            radius: Tokens.radius
                            color: active ? Tokens.tint10 : "transparent"
                            border.width: active ? Tokens.border : 0
                            border.color: Tokens.line

                            Row {
                                anchors.fill: parent
                                anchors.leftMargin: Tokens.s3
                                anchors.rightMargin: Tokens.s3
                                spacing: Tokens.s3

                                Rectangle {
                                    width: 22; height: 22
                                    radius: 11
                                    anchors.verticalCenter: parent.verticalCenter
                                    color: done ? Tokens.ink : (active ? Tokens.ink : "transparent")
                                    border.width: Tokens.border
                                    border.color: (done || active) ? Tokens.ink : Tokens.line

                                    Text {
                                        anchors.centerIn: parent
                                        text: done ? "\u2713" : String(index + 1)
                                        color: (done || active) ? Tokens.paper : Tokens.inkMuted
                                        font.family: Tokens.mono
                                        font.pixelSize: 11
                                        font.weight: Font.Bold
                                    }
                                }

                                Text {
                                    anchors.verticalCenter: parent.verticalCenter
                                    text: modelData.title
                                    color: active ? Tokens.ink : (done ? Tokens.inkMuted : Tokens.inkFaint)
                                    font.family: Tokens.ui
                                    font.pixelSize: Tokens.fSmall
                                    font.weight: active ? Font.Medium : Font.Normal
                                }
                            }
                        }
                    }
                }
            }

            Text {
                anchors.left: parent.left; anchors.right: parent.right; anchors.bottom: parent.bottom
                anchors.margins: Tokens.s6
                text: I18n.tr("Nothing is written to your drive until you click Install Now.")
                color: Tokens.inkFaint
                font.family: Tokens.ui
                font.pixelSize: Tokens.fMicro
                wrapMode: Text.WordWrap
            }
        }

        // Right main pane
        Item {
            width: parent.width - 248
            height: parent.height

            // Header
            Column {
                id: pageHead
                anchors.left: parent.left; anchors.right: parent.right; anchors.top: parent.top
                anchors.margins: Tokens.s6
                spacing: Tokens.s1

                Text {
                    text: root.steps[root.step].title
                    color: Tokens.ink
                    font.family: Tokens.display
                    font.pixelSize: Tokens.fTitle
                }
                Text {
                    width: parent.width
                    text: root.steps[root.step].desc
                    color: Tokens.inkMuted
                    font.family: Tokens.ui
                    font.pixelSize: Tokens.fBody
                    wrapMode: Text.WordWrap
                }
            }

            // Content body
            Flickable {
                id: bodyFlick
                anchors.left: parent.left; anchors.right: parent.right
                anchors.top: pageHead.bottom; anchors.bottom: footerBar.top
                anchors.margins: Tokens.s6
                contentWidth: width
                contentHeight: stepContent.implicitHeight + Tokens.s4
                clip: true
                boundsBehavior: Flickable.StopAtBounds

                Column {
                    id: stepContent
                    width: bodyFlick.width
                    spacing: Tokens.s5

                    // ── STEP 0: REGION ─────────────────────────────────────
                    Column {
                        width: parent.width
                        spacing: Tokens.s4
                        visible: root.step === 0

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("LANGUAGE & LOCALE"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            Flow {
                                width: parent.width
                                spacing: Tokens.s2
                                Repeater {
                                    model: root.locales
                                    delegate: ChoiceChip {
                                        required property var modelData
                                        text: modelData.label + " (" + modelData.code + ")"
                                        selected: root.locale === modelData.code
                                        onClicked: root.locale = modelData.code
                                    }
                                }
                            }
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("KEYBOARD LAYOUT"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            Flow {
                                width: parent.width
                                spacing: Tokens.s2
                                Repeater {
                                    model: root.keymaps
                                    delegate: ChoiceChip {
                                        required property string modelData
                                        text: modelData.toUpperCase()
                                        selected: root.keymap === modelData
                                        onClicked: root.keymap = modelData
                                    }
                                }
                            }
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("TIMEZONE"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            Flow {
                                width: parent.width
                                spacing: Tokens.s2
                                Repeater {
                                    model: root.timezones
                                    delegate: ChoiceChip {
                                        required property string modelData
                                        text: modelData
                                        selected: root.timezone === modelData
                                        onClicked: root.timezone = modelData
                                    }
                                }
                            }
                            FieldBox {
                                width: parent.width
                                placeholderText: I18n.tr("Or enter IANA timezone (e.g. Europe/Paris)")
                                text: root.timezone
                                onTextEdited: root.timezone = text
                            }
                        }
                    }

                    // ── STEP 1: DISK & STORAGE ─────────────────────────────
                    Column {
                        width: parent.width
                        spacing: Tokens.s4
                        visible: root.step === 1

                        Row {
                            width: parent.width
                            Text {
                                width: parent.width - refreshBtn.width
                                anchors.verticalCenter: parent.verticalCenter
                                text: I18n.tr("TARGET DRIVE")
                                color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel
                            }
                            ActionBtn {
                                id: refreshBtn
                                text: I18n.tr("Refresh")
                                onClicked: root.refreshDisks()
                            }
                        }

                        Text {
                            visible: root.disksLoading
                            text: I18n.tr("Scanning drives...")
                            color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fSmall
                        }

                        Text {
                            visible: !root.disksLoading && (root.disks.length === 0 || root.diskError !== "")
                            text: root.diskError !== "" ? root.diskError : I18n.tr("No physical drives (>= 4 GB) detected.")
                            color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fSmall
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Repeater {
                                model: root.disks
                                delegate: Rectangle {
                                    required property var modelData
                                    readonly property bool sel: root.selectedDisk === modelData.path
                                    width: parent.width
                                    height: 54
                                    radius: Tokens.radius
                                    color: sel ? Tokens.tint10 : Tokens.tint5
                                    border.width: sel ? 2 : Tokens.border
                                    border.color: sel ? Tokens.ink : Tokens.line

                                    Row {
                                        anchors.fill: parent
                                        anchors.leftMargin: Tokens.s4
                                        anchors.rightMargin: Tokens.s4
                                        spacing: Tokens.s3

                                        Column {
                                            width: parent.width - sizeLbl.width - Tokens.s3
                                            anchors.verticalCenter: parent.verticalCenter
                                            spacing: 2
                                            Text {
                                                text: modelData.path + (modelData.model ? ("  \u00b7  " + modelData.model) : "")
                                                color: Tokens.ink; font.family: Tokens.mono; font.pixelSize: Tokens.fSmall; font.weight: Font.Medium
                                            }
                                            Text {
                                                text: modelData.tran ? modelData.tran : I18n.tr("Local disk")
                                                color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro
                                            }
                                        }

                                        Text {
                                            id: sizeLbl
                                            anchors.verticalCenter: parent.verticalCenter
                                            text: modelData.sizeText
                                            color: Tokens.ink; font.family: Tokens.mono; font.pixelSize: Tokens.fSmall
                                        }
                                    }

                                    MouseArea {
                                        anchors.fill: parent
                                        cursorShape: Qt.PointingHandCursor
                                        onClicked: root.selectedDisk = modelData.path
                                    }
                                }
                            }
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("INSTALLATION TYPE"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            Row {
                                width: parent.width
                                spacing: Tokens.s2
                                ChoiceChip {
                                    text: I18n.tr("Erase disk and install Toren")
                                    selected: root.installMode === "wipe"
                                    onClicked: root.installMode = "wipe"
                                }
                                ChoiceChip {
                                    text: I18n.tr("Install alongside existing OS")
                                    selected: root.installMode === "alongside"
                                    onClicked: root.installMode = "alongside"
                                }
                            }
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("FILESYSTEM & ENCRYPTION"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            Row {
                                spacing: Tokens.s2
                                ChoiceChip { text: "Btrfs (Recommended)"; selected: root.filesystem === "btrfs"; onClicked: root.filesystem = "btrfs" }
                                ChoiceChip { text: "ext4"; selected: root.filesystem === "ext4"; onClicked: root.filesystem = "ext4" }
                                ChoiceChip {
                                    text: root.encrypt ? I18n.tr("LUKS2 Encryption: On") : I18n.tr("LUKS2 Encryption: Off")
                                    selected: root.encrypt
                                    onClicked: root.encrypt = !root.encrypt
                                }
                            }
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            visible: root.encrypt
                            FieldBox {
                                width: parent.width
                                echoMode: TextInput.Password
                                placeholderText: I18n.tr("Encryption passphrase (min 4 characters)")
                                text: root.luksPass
                                onTextEdited: root.luksPass = text
                            }
                            FieldBox {
                                width: parent.width
                                echoMode: TextInput.Password
                                placeholderText: I18n.tr("Confirm encryption passphrase")
                                text: root.luksConfirm
                                onTextEdited: root.luksConfirm = text
                            }
                            Text {
                                visible: root.luksPass.length > 0 && root.luksPass !== root.luksConfirm
                                text: I18n.tr("Passphrases do not match.")
                                color: Tokens.alert; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro
                            }
                        }
                    }

                    // ── STEP 2: ACCOUNT ────────────────────────────────────
                    Column {
                        width: parent.width
                        spacing: Tokens.s4
                        visible: root.step === 2

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("USERNAME"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            FieldBox {
                                width: parent.width
                                placeholderText: "toren"
                                text: root.username
                                onTextEdited: root.username = text
                            }
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("COMPUTER NAME (HOSTNAME)"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            FieldBox {
                                width: parent.width
                                placeholderText: "toren"
                                text: root.hostname
                                onTextEdited: root.hostname = text
                            }
                        }

                        Column {
                            width: parent.width
                            spacing: Tokens.s2
                            Text { text: I18n.tr("PASSWORD"); color: Tokens.inkMuted; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro; font.letterSpacing: Tokens.trackLabel }
                            FieldBox {
                                width: parent.width
                                echoMode: TextInput.Password
                                placeholderText: I18n.tr("Enter password")
                                text: root.password
                                onTextEdited: root.password = text
                            }
                            FieldBox {
                                width: parent.width
                                echoMode: TextInput.Password
                                placeholderText: I18n.tr("Confirm password")
                                text: root.passwordConfirm
                                onTextEdited: root.passwordConfirm = text
                            }
                            Text {
                                visible: root.password.length > 0 && root.password !== root.passwordConfirm
                                text: I18n.tr("Passwords do not match.")
                                color: Tokens.alert; font.family: Tokens.ui; font.pixelSize: Tokens.fMicro
                            }
                        }
                    }

                    // ── STEP 3: REVIEW ─────────────────────────────────────
                    Column {
                        width: parent.width
                        spacing: Tokens.s4
                        visible: root.step === 3

                        Rectangle {
                            width: parent.width
                            height: summaryCol.implicitHeight + Tokens.s5 * 2
                            radius: Tokens.radius
                            color: Tokens.tint5
                            border.width: Tokens.border
                            border.color: Tokens.line

                            Column {
                                id: summaryCol
                                anchors.fill: parent
                                anchors.margins: Tokens.s5
                                spacing: Tokens.s3

                                SummaryRow { label: I18n.tr("Locale / Keyboard"); value: root.locale + "  \u00b7  " + root.keymap }
                                SummaryRow { label: I18n.tr("Timezone"); value: root.timezone }
                                SummaryRow { label: I18n.tr("Target Drive"); value: root.selectedDisk }
                                SummaryRow { label: I18n.tr("Mode"); value: root.installMode === "wipe" ? I18n.tr("Erase disk & clean install") : I18n.tr("Install alongside existing OS") }
                                SummaryRow { label: I18n.tr("Filesystem"); value: root.filesystem + (root.encrypt ? " + LUKS2" : "") }
                                SummaryRow { label: I18n.tr("User & Hostname"); value: root.username + "@" + root.hostname }
                            }
                        }

                        Rectangle {
                            visible: root.installMode === "wipe"
                            width: parent.width
                            height: warnText.implicitHeight + Tokens.s4 * 2
                            radius: Tokens.radius
                            color: Qt.rgba(Tokens.alert.r, Tokens.alert.g, Tokens.alert.b, 0.12)
                            border.width: Tokens.border
                            border.color: Qt.rgba(Tokens.alert.r, Tokens.alert.g, Tokens.alert.b, 0.45)

                            Text {
                                id: warnText
                                anchors.fill: parent
                                anchors.margins: Tokens.s4
                                text: I18n.tr("Warning: All existing partitions and data on %1 will be permanently erased.").arg(root.selectedDisk)
                                color: Tokens.ink
                                font.family: Tokens.ui
                                font.pixelSize: Tokens.fSmall
                                wrapMode: Text.WordWrap
                            }
                        }
                    }

                    // ── STEP 4: INSTALLING ─────────────────────────────────
                    Column {
                        width: parent.width
                        spacing: Tokens.s4
                        visible: root.step === 4

                        Text {
                            text: root.installStatus
                            color: Tokens.ink
                            font.family: Tokens.ui
                            font.pixelSize: Tokens.fBody
                            font.weight: Font.Medium
                        }

                        Rectangle {
                            width: parent.width
                            height: 8
                            radius: 4
                            color: Tokens.tint10

                            Rectangle {
                                width: parent.width * (root.installProgress / 100)
                                height: parent.height
                                radius: 4
                                color: root.installDone && !root.installOk ? Tokens.alert : Tokens.ink
                                Behavior on width { NumberAnimation { duration: 240; easing.type: Easing.OutCubic } }
                            }
                        }

                        Rectangle {
                            width: parent.width
                            height: 290
                            radius: Tokens.radius
                            color: Tokens.tint5
                            border.width: Tokens.border
                            border.color: Tokens.line

                            ScrollView {
                                anchors.fill: parent
                                anchors.margins: Tokens.s3
                                clip: true
                                TextArea {
                                    readOnly: true
                                    text: root.installLog
                                    color: Tokens.inkMuted
                                    font.family: Tokens.mono
                                    font.pixelSize: 11
                                    wrapMode: TextArea.WrapAnywhere
                                    background: null
                                }
                            }
                        }
                    }
                }
            }

            // Bottom action bar
            Rectangle {
                id: footerBar
                anchors.left: parent.left; anchors.right: parent.right; anchors.bottom: parent.bottom
                height: 68
                color: Tokens.paper

                Rectangle {
                    anchors.top: parent.top
                    width: parent.width
                    height: Tokens.border
                    color: Tokens.line
                }

                Row {
                    anchors.right: parent.right
                    anchors.verticalCenter: parent.verticalCenter
                    anchors.rightMargin: Tokens.s6
                    spacing: Tokens.s3

                    ActionBtn {
                        visible: root.step > 0 && root.step < 4
                        text: I18n.tr("Back")
                        onClicked: root.step -= 1
                    }
                    ActionBtn {
                        visible: root.step < 3
                        primary: true
                        enabled: root.canAdvance
                        text: I18n.tr("Continue")
                        onClicked: root.step += 1
                    }
                    ActionBtn {
                        visible: root.step === 3
                        primary: true
                        enabled: root.canAdvance
                        text: I18n.tr("Install Now")
                        onClicked: root.startInstall()
                    }
                    ActionBtn {
                        visible: root.step === 4 && root.installDone && !root.installOk
                        text: I18n.tr("Back to Review")
                        onClicked: root.step = 3
                    }
                    ActionBtn {
                        visible: root.step === 4 && root.installDone && root.installOk
                        text: I18n.tr("Continue Testing")
                        onClicked: Qt.quit()
                    }
                    ActionBtn {
                        visible: root.step === 4 && root.installDone && root.installOk
                        primary: true
                        text: I18n.tr("Reboot Now")
                        onClicked: rebootProc.running = true
                    }
                }
            }
        }
    }

    // ── local reusable controls ─────────────────────────────────────────────
    component ChoiceChip: Rectangle {
        id: chip
        property string text: ""
        property bool selected: false
        signal clicked()

        implicitWidth: chipLbl.implicitWidth + 24
        implicitHeight: 34
        radius: Tokens.radius
        color: selected ? Tokens.ink : (chipMa.containsMouse ? Tokens.tint10 : Tokens.tint5)
        border.width: Tokens.border
        border.color: selected ? Tokens.ink : Tokens.line

        Text {
            id: chipLbl
            anchors.centerIn: parent
            text: chip.text
            color: chip.selected ? Tokens.paper : Tokens.ink
            font.family: Tokens.ui
            font.pixelSize: Tokens.fSmall
            font.weight: chip.selected ? Font.Medium : Font.Normal
        }

        MouseArea {
            id: chipMa
            anchors.fill: parent
            hoverEnabled: true
            cursorShape: Qt.PointingHandCursor
            onClicked: chip.clicked()
        }
    }

    component ActionBtn: Rectangle {
        id: abtn
        property string text: ""
        property bool primary: false
        signal clicked()

        implicitWidth: abtnLbl.implicitWidth + 28
        implicitHeight: 36
        radius: Tokens.radius
        opacity: enabled ? 1.0 : 0.45
        color: primary ? Tokens.ink : (abtnMa.containsMouse ? Tokens.tint10 : Tokens.tint5)
        border.width: Tokens.border
        border.color: primary ? Tokens.ink : Tokens.line

        Text {
            id: abtnLbl
            anchors.centerIn: parent
            text: abtn.text
            color: abtn.primary ? Tokens.paper : Tokens.ink
            font.family: Tokens.ui
            font.pixelSize: Tokens.fSmall
            font.weight: Font.Medium
        }

        MouseArea {
            id: abtnMa
            anchors.fill: parent
            enabled: abtn.enabled
            hoverEnabled: true
            cursorShape: abtn.enabled ? Qt.PointingHandCursor : Qt.ArrowCursor
            onClicked: abtn.clicked()
        }
    }

    component FieldBox: TextField {
        color: Tokens.ink
        placeholderTextColor: Tokens.inkFaint
        font.family: Tokens.ui
        font.pixelSize: Tokens.fSmall
        leftPadding: 12
        rightPadding: 12
        implicitHeight: 38
        background: Rectangle {
            radius: Tokens.radius
            color: Tokens.tint5
            border.width: activeFocus ? 2 : Tokens.border
            border.color: activeFocus ? Tokens.ink : Tokens.line
        }
    }

    component SummaryRow: Row {
        property string label: ""
        property string value: ""
        width: parent.width
        Text {
            width: 170
            text: parent.label
            color: Tokens.inkMuted
            font.family: Tokens.ui
            font.pixelSize: Tokens.fSmall
        }
        Text {
            width: parent.width - 170
            text: parent.value
            color: Tokens.ink
            font.family: Tokens.mono
            font.pixelSize: Tokens.fSmall
            elide: Text.ElideRight
        }
    }
}
