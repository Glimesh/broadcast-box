import { localeInterface } from "./localeInterface"

const locale_en: localeInterface = {
  locale: "en",
  frontpage: {
    welcome: "Welcome to Kohi-Box",
    welcome_subtitle: "Kohi-Box is a tool that allows you to efficiently stream high-quality video in real time, using the latest in video codecs and WebRTC technology.",

    toggle_watch: "I want to watch",
    toggle_stream: "I want to stream",

    stream_input_label: "Stream key",
    stream_input_placeholder_share: "paste the private key of the stream you want to share",
    stream_input_placeholder_join: "paste the key of the stream you want to join",

    stream_button_stream_start: "start stream ♡",
    stream_button_stream_join: "join stream ♡",
  },

  available_streams: {
    title: "Current streams",
    stream_join_message: "Click a stream to join it",
    no_streams_message: "no streams right now — check back soon ♡",
  },

  statistics: {
    title: "Statistics",
    no_statistics_available: "nothing live right now — check back soon ♡",
    no_sessions: "no sessions available",

    whep_sessions: "WHEP Sessions",

    video: "Video",
    video_tracks: "Video tracks",
    video_track_not_available: "No video tracks",
    video_bitrate: "Bitrate",
    video_bitrate_total_out: "Bitrate outgoing",

    audio: "Audio",
    audio_tracks: "Audio tracks",
    audio_track_not_available: "No audio tracks",

    button_watch_stream: "Watch stream",

    rid: "RID",
    layer: "Layers",
    packets_received: "Packets received",
    packets_dropped: "Packets dropped",
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
    mediaAccessError_noMediaDevices: "MediaDevices API was not found. Publishing in Kohi-Box requires HTTPS",
    mediaAccessError_notAllowedError: "You can't publish stream using your camera, access has been disabled.",
    mediaAccessError_notFoundError: "Seems like you don't have camera. Or the access to it is blocked\nCheck camera settings, browser permissions and system permissions.",

    connection_established: "Live: Currently streaming to",
    connection_disconnected: "WebRTC has disconnected or failed to connect at all",
    connection_failed: "Failed to start Kohi-Box session",
    connection_has_packetloss: "WebRTC is experiencing packet loss",

    publish_screen: "publish screen/window/tab",
    publish_webcam: "publish webcam",

    button_end_stream: "end stream"
  },

  player_page: {
    cinema_mode_disable: "disable cinema mode",
    cinema_mode_enable: "enable cinema mode",

    modal_add_stream_title: "add stream",
    modal_add_stream_message: "paste a stream key to add it to the multi-stream view",
    modal_add_stream_placeholder: "paste the key of the stream you want to add",

    visibility_public: "set private",
    visibility_private: "set public",

    edit_stream_title: "edit stream title",
    modal_edit_stream_title: "edit stream title",
    modal_edit_stream_message: "this is the little name shown above your stream",
    modal_edit_stream_placeholder: "enter a title for your stream",
    error_edit_stream_title: "couldn't save the title, try again ♡",
    error_edit_stream_empty: "give your stream a name first ♡",
  },

  player: {
    message_is_not_online: "isn't streaming right now",
    message_loading_video: "loading video...",
    message_error: "couldn't load the video",

    stream_status_offline: "offline"
  },

  stream_status: {
    message_current_viewers: "current viewers"
  },

  profile_settings: {
    title: "profile settings",
    subTitle: "configure your stream profile",

    toggle_stream_privacy_label: "privacy",
    toggle_stream_privacy_title_left: "private",
    toggle_stream_privacy_title_right: "public",

    input_motd_label: "message of the day",
    button_save_label: "save ♡"
  },

  admin_login: {
    login_input_dialog_title: "admin login",
    login_input_dialog_message: "paste your admin token to sign in",
    login_input_dialog_placeholder: "paste your admin token",
    error_message_login_failed: "login failed",

    button_login_text: "sign in ♡"
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

    yes: "Yes",
    no: "No"
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
    button_accept: "ok ♡"
  },
  shared_component_text_input_modal: {
    button_accept: "ok ♡"
  },
  shared_component_text_input_dialog: {
    button_accept: "ok ♡"
  },

  chat: {
    title: "chat",
    placeholder_input: "say something nice...",
    button_change_display_name_title: "change display name",
    button_send_title: "send message",
    status_connecting: "connecting",
    status_connected: "connected",
    status_error: "error",
    status_disconnected: "disconnected",
    no_messages_yet: "no messages yet — say hi ♡",
    error_failed_to_connect: "couldn't connect to chat",
    error_not_connected: "chat is not connected",
    error_failed_to_send: "couldn't send message",
    modal_display_name_title: "display name",
    modal_display_name_message: "pick a name for chat",
    modal_display_name_placeholder: "your name",
  },
}
export default locale_en
