import { localeInterface } from "./localeInterface"

const locale_en: localeInterface = {
  locale: "en",
  frontpage: {
    welcome: "Welcome to Broadcast Box",
    welcome_subtitle: "Broadcast Box is a tool that allows you to efficiently stream high-quality video in real time, using the latest in video codecs and WebRTC technology.",

    toggle_watch: "I want to watch",
    toggle_stream: "I want to stream",

    stream_input_label: "Stream key",
    stream_input_placeholder_share: "Insert the key of the stream you want to share",
    stream_input_placeholder_join: "Insert the key of the stream you want to join",

    stream_button_stream_start: "Start stream",
    stream_button_stream_join: "Join stream",
  },

  available_streams: {
    title: "Current streams",
    stream_join_message: "Click a stream to join it",
    no_streams_message: "No streams currently available",
  },

  statistics: {
    title: "Statistics",
    no_statistics_available: "No statistics currently available",
    no_sessions: "No sessions available",

    whep_sessions: "WHEP Sessions",

    video: "Video",
    video_tracks: "Video tracks",
    video_track_not_available: "No video tracks",

    audio: "Video",
    audio_tracks: "Audio tracks",
    audio_track_not_available: "No audio tracks",

    button_watch_stream: "Watch stream",

    rid: "RID",
    layer: "Layers",
    packets_received: "Packets received",
    packets_written: "Packets written",
    last_key_frame: "Last key frame",
    timestamp: "Timestamp",
    sequence_number: "Seq. number"
  },

  player_header: {
    error: "Error",
    success: "Success",
    warning: "Warning",

    mediaAccessError_default: "Could not access your media devices",
    mediaAccessError_noMediaDevices: "MediaDevices API was not found. Publishing in Broadcast Box requires HTTPS",
    mediaAccessError_notAllowedError: "You can't publish stream using your camera, access has been disabled.",
    mediaAccessError_notFoundError: "Seems like you don't have camera. Or the access to it is blocked\nCheck camera settings, browser permissions and system permissions.",

    connection_established: "Live: Currently streaming to",
    connection_disconnected: "WebRTC has disconnected or failed to connect at all",
    connection_failed: "Failed to start Broadcast Box session",
    connection_has_packetloss: "WebRTC is experiencing packet loss",

    publish_screen: "Publish Screen/Window/Tab",
    publish_webcam: "Publish Webcam",

    button_end_stream: "End stream"
  },

  player_page: {
    cinema_mode_disable: "Disable cinema mode",
    cinema_mode_enable: "Enable cinema mode",

    modal_add_stream_title: "Add stream",
    modal_add_stream_message: "Insert stream key to add to multi stream",
    modal_add_stream_placeholder: "Insert the key of the stream you want to add",

    stream_status_offline: "Offline"
  },

  player: {
    message_is_not_online: "is not currently streaming",
    message_loading_video: "Loading video"
  },

  stream_status: {
    message_current_viewers: "Current Viewers"
  },

  profile_settings: {
    title: "Profile Settings",
    subTitle: "Configure streaming profile",

    toggle_stream_privacy_label: "Privacy",
    toggle_stream_privacy_title_left: "Private",
    toggle_stream_privacy_title_right: "Public",

    input_motd_label: "Message of the Day",
    button_save_label: "Save"
  },

  admin_login: {
    login_input_dialog_title: "Login",
    login_input_dialog_message: "Insert admin token to log in",
    login_input_dialog_placeholder: "Insert admin token to log in",
    error_message_login_failed: "Login failed"
  },

  admin_page: {
    title: "Admin Portal",
    menu_api: "API",
    menu_logging: "Logging",
    menu_logout: "Logout",
    menu_profiles: "Profiles",
    menu_settings: "Settings",
    menu_status: "Status"
  },

  admin_page_api: {
    title: "API Settings",
    table_header_setting_name: "Setting",
    table_header_value: "Value"
  },

  admin_page_logging: {
    title: "Logging",
    table_header_setting_name: "Setting",
    table_header_value: "Value"
  },

  admin_page_profiles: {
    title: "Profiles Overview",

    add_profile_modal_title: "Add Profile",
    add_profile_modal_message: "Insert a key to add a new stream",
    add_profile_modal_placeholder: "Write new stream key here",

    remove_profile_modal_title: "Remove Profile",
    remove_profile_modal_message: "Are you sure you would like to remove",

    table_header_stream_key: "Stream Key",
    table_header_is_public: "Is Public",
    table_header_motd: "Motd",
    table_header_token: "Token",

    button_add_profile: "Add profile",
  },

  admin_page_status_page: {
    title: "Stream Status Overview",

    table_header_stream_key: "Stream Key",
    table_header_is_public: "Is Public",
    table_header_video_tracks: "Video Tracks",
    table_header_audio_tracks: "Audio Tracks",
    table_header_sessions: "Sessions",
    table_header_total_packets: "Total Packets",

    yes: "yes",
    no: "no"
  },

  shared_component_card: {
    button_accept: "Accept"
  },
  shared_component_text_input_modal: {
    button_accept: "Accept"
  },
  shared_component_text_input_dialog: {
    button_accept: "Accept"
  },
}
export default locale_en
