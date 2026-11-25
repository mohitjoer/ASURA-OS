pragma Singleton
pragma ComponentBehavior: Bound

// From https://git.outfoxxed.me/outfoxxed/nixnew
// It does not have a license, but the author is okay with redistribution.

import QtQml.Models
import QtQuick
import Quickshell
import Quickshell.Io
import Quickshell.Services.Mpris
import qs.modules.common

/**
 * A service that provides easy access to the active Mpris player.
 */
Singleton {
	id: root

	function pickActive(players, cur) {
		var list = (players || []).filter(function(p) { return p && isRealPlayer(p); });
		if (list.length === 0) return null;
		var curOk = cur && list.indexOf(cur) !== -1;
		if (curOk && cur.isPlaying) return cur;
		for (var i = 0; i < list.length; i++) {
			if (list[i].isPlaying || list[i].playbackState === MprisPlaybackState.Playing)
				return list[i];
		}
		if (curOk && (cur.trackTitle ?? "").length > 0) return cur;
		for (var j = 0; j < list.length; j++) {
			if ((list[j].trackTitle ?? "").length > 0) return list[j];
		}
		return curOk ? cur : list[0];
	}

	function repick() {
		var next = pickActive(Mpris.players ? Mpris.players.values : [], root.trackedPlayer);
		if (next !== root.trackedPlayer) {
			root.trackedPlayer = next;
		}
	}

	Timer {
		interval: 1000
		repeat: true
		running: true
		onTriggered: root.repick()
	}

	Component.onCompleted: root.repick()

	property list<MprisPlayer> players: Mpris.players.values.filter(player => isRealPlayer(player))
	property MprisPlayer trackedPlayer: null
	property MprisPlayer activePlayer: trackedPlayer ?? pickActive(Mpris.players ? Mpris.players.values : [], null)
	signal trackChanged(reverse: bool)

	property bool __reverse: false

	property var activeTrack

	readonly property bool hasActivePlasmaIntegration: Mpris.players.values.some(
		p => p.dbusName?.startsWith('org.mpris.MediaPlayer2.plasma-browser-integration')
	)

	function isRealPlayer(player) {
		if (!player) return false;
		if (!Config.options.media.filterDuplicatePlayers) {
			return true;
		}
		var name = String(player.dbusName || "");
		return (
			// Remove native browser buses only if plasma-browser-integration is actually active on D-Bus
			!(hasActivePlasmaIntegration && (name.startsWith('org.mpris.MediaPlayer2.firefox') || name.startsWith('org.mpris.MediaPlayer2.chromium'))) &&
			// playerctld just copies other buses and we don't need duplicates
			!name.startsWith('org.mpris.MediaPlayer2.playerctld') &&
			// Non-instance mpd bus
			!(name.endsWith('.mpd') && !name.endsWith('MediaPlayer2.mpd'))
		);
	}

	Instantiator {
		model: Mpris.players

		Connections {
			required property MprisPlayer modelData
			target: modelData

			Component.onCompleted: {
				if (root.trackedPlayer == null || modelData.isPlaying) {
					root.trackedPlayer = modelData;
				}
			}

			Component.onDestruction: {
				if (root.trackedPlayer == null || !root.trackedPlayer.isPlaying) {
					for (const player of Mpris.players.values) {
						if (player && (player.isPlaying || player.playbackState === MprisPlaybackState.Playing)) {
							root.trackedPlayer = player;
							break;
						}
					}

					if (root.trackedPlayer == null && Mpris.players.values.length != 0) {
						root.trackedPlayer = Mpris.players.values[0];
					}
				}
			}

			function onPlaybackStateChanged() {
				if (modelData.isPlaying) root.trackedPlayer = modelData;
			}
		}
	}

	Connections {
		target: activePlayer

		function onPostTrackChanged() {
			root.updateTrack();
		}

		function onTrackArtUrlChanged() {
			if (root.activePlayer && root.activeTrack && root.activePlayer.uniqueId == root.activeTrack.uniqueId && root.activePlayer.trackArtUrl != root.activeTrack.artUrl) {
				const r = root.__reverse;
				root.updateTrack();
				root.__reverse = r;
			}
		}
	}

	onActivePlayerChanged: this.updateTrack();

	function updateTrack() {
		this.activeTrack = {
			uniqueId: this.activePlayer?.uniqueId ?? 0,
			artUrl: this.activePlayer?.trackArtUrl ?? "",
			title: this.activePlayer?.trackTitle || Translation.tr("Unknown Title"),
			artist: this.activePlayer?.trackArtist || Translation.tr("Unknown Artist"),
			album: this.activePlayer?.trackAlbum || Translation.tr("Unknown Album"),
		};

		this.trackChanged(__reverse);
		this.__reverse = false;
	}

	property bool isPlaying: this.activePlayer && this.activePlayer.isPlaying;
	property bool canTogglePlaying: this.activePlayer?.canTogglePlaying ?? false;
	function togglePlaying() {
		if (this.canTogglePlaying) this.activePlayer.togglePlaying();
	}

	property bool canGoPrevious: this.activePlayer?.canGoPrevious ?? false;
	function previous() {
		if (this.canGoPrevious) {
			this.__reverse = true;
			this.activePlayer.previous();
		}
	}

	property bool canGoNext: this.activePlayer?.canGoNext ?? false;
	function next() {
		if (this.canGoNext) {
			this.__reverse = false;
			this.activePlayer.next();
		}
	}

	property bool canChangeVolume: this.activePlayer && this.activePlayer.volumeSupported && this.activePlayer.canControl;

	property bool loopSupported: this.activePlayer && this.activePlayer.loopSupported && this.activePlayer.canControl;
	property var loopState: this.activePlayer?.loopState ?? MprisLoopState.None;
	function setLoopState(loopState: var) {
		if (this.loopSupported) {
			this.activePlayer.loopState = loopState;
		}
	}

	property bool shuffleSupported: this.activePlayer && this.activePlayer.shuffleSupported && this.activePlayer.canControl;
	property bool hasShuffle: this.activePlayer?.shuffle ?? false;
	function setShuffle(shuffle: bool) {
		if (this.shuffleSupported) {
			this.activePlayer.shuffle = shuffle;
		}
	}

	function setActivePlayer(player: MprisPlayer) {
		const targetPlayer = player ?? (Mpris.players.values.length > 0 ? Mpris.players.values[0] : null);
		if (targetPlayer && this.activePlayer) {
			this.__reverse = Mpris.players.indexOf(targetPlayer) < Mpris.players.indexOf(this.activePlayer);
		} else {
			this.__reverse = false;
		}
		this.trackedPlayer = targetPlayer;
	}

	IpcHandler {
		target: "mpris"

		function pauseAll(): void {
			for (const player of Mpris.players.values) {
				if (player.canPause) player.pause();
			}
		}

		function playPause(): void { root.togglePlaying(); }
		function previous(): void { root.previous(); }
		function next(): void { root.next(); }
	}
}
