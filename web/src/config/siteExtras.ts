const videoIds = [
	"v3gOQbh32Pg",
	"0yiwxIuXmdk",
	"ZUgG3LGGbb4",
	"frIxIZyko80",
	"M6CmkoF-1oM",
];

export interface MusicPlaylist {
	title: string;
	subtitle: string;
	videoIds: string[];
	defaultVolume: number;
}

export const DEFAULT_MUSIC_PLAYLIST: MusicPlaylist = {
	title: "frutiger aero radio",
	subtitle: "nintendo mix · eShop / Wii / 3DS / DS",
	videoIds,
	defaultVolume: 30,
};

export const MUSIC_PLAYLIST = {
	...DEFAULT_MUSIC_PLAYLIST,
	embedVideoId: DEFAULT_MUSIC_PLAYLIST.videoIds[0],
	youtubeUrl: `https://www.youtube.com/watch_videos?video_ids=${encodeURIComponent(DEFAULT_MUSIC_PLAYLIST.videoIds.join(","))}`,
};

// v2 keys intentionally start fresh so the old 0% volume bug cannot stick.
export const MUSIC_STORAGE_KEY = "bb_music_enabled_v2";
export const MUSIC_VOLUME_STORAGE_KEY = "bb_music_volume_v2";
export const MUSIC_PROMPTED_STORAGE_KEY = "bb_music_prompted_v2";
export const THEME_STORAGE_KEY = "bb_theme";
