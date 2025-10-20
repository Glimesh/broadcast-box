export type localeTypes = "en" | "da"

export interface localeInterface {
  locale: localeTypes,
  frontpage: {
    welcome: string,
    welcome_subtitle: string,

    toggle_watch: string,
    toggle_stream: string,

    stream_input_label: string,
    stream_input_placeholder_share: string,
    stream_input_placeholder_join: string,

    stream_button_stream_start: string,
    stream_button_stream_join: string
  },

  available_streams: {
    title: string,
    stream_join_message: string
    no_streams_message: string
  },

  statistics: {
    title: string,
    no_statistics_available: string,
    no_sessions: string,

    whep_sessions: string,

    video: string,
    video_tracks: string,
    video_track_not_available: string,

    audio: string,
    audio_tracks: string,
    audio_track_not_available: string,

    button_watch_stream: string,

    rid: string,
    layer: string,
    packets_received: string,
    packets_written: string,
    last_key_frame: string
    timestamp: string
    sequence_number: string
  },

  player_header: {
    error: string,
    warning: string,
    success: string,

    mediaAccessError_noMediaDevices: string,
    mediaAccessError_notFoundError: string,
    mediaAccessError_notAllowedError: string,
    mediaAccessError_default: string,

    connection_established: string,
    connection_disconnected: string,
    connection_failed: string,
    connection_has_packetloss: string,

    publish_screen: string,
    publish_webcam: string,

    button_end_stream: string,
  },

  player_page: {
    modal_add_stream_title: string,
    modal_add_stream_message: string,
    modal_add_stream_placeholder: string,

    cinema_mode_enable: string,
    cinema_mode_disable: string,
  },
  player: {
    message_is_not_online: string,
    message_loading_video: string,
    message_error: string,

    stream_status_offline: string,
  },
  stream_status: {
    message_current_viewers: string
  },

  profile_settings: {
    title: string,
    subTitle: string,

    input_motd_label: string,

    toggle_stream_privacy_label: string
    toggle_stream_privacy_title_left: string
    toggle_stream_privacy_title_right: string

    button_save_label: string
  },

  shared_component_text_input_dialog: {
    button_accept: string;
  },

  shared_component_text_input_modal: {
    button_accept: string;
  },

  shared_component_card: {
    button_accept: string;
  },

  admin_login: {
    login_input_dialog_title: string,
    login_input_dialog_message: string,
    login_input_dialog_placeholder: string,

    error_message_login_failed: string

    button_login_text: string
  },

  admin_page: {
    title: string,

    menu_status: string,
    menu_profiles: string,
    menu_api: string,
    menu_logging: string,
    menu_settings: string,
    menu_logout: string
  },

  admin_page_api: {
    title: string,

    table_header_setting_name: string,
    table_header_value: string,
  },

  admin_page_logging: {
    title: string,

    table_header_setting_name: string,
    table_header_value: string,
  },

  admin_page_profiles: {
    title: string,

    add_profile_modal_title: string,
    add_profile_modal_message: string,
    add_profile_modal_placeholder: string,

    remove_profile_modal_title: string,
    remove_profile_modal_message: string,

    table_header_stream_key: string,
    table_header_is_public: string,
    table_header_motd: string,
    table_header_token: string,

    button_add_profile: string,

    yes: string,
    no: string
  },

  admin_page_status_page: {
    title: string,

    table_header_stream_key: string,
    table_header_is_public: string,
    table_header_video_tracks: string,
    table_header_audio_tracks: string,
    table_header_sessions: string,
    table_header_total_packets: string,

    yes: string,
    no: string

  }
}
