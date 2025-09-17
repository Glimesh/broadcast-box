import { localeInterface } from "./localeInterface"

const locale_da: localeInterface = {
  locale: "da",
  frontpage: {
    welcome: "Velkommen til Broadcast Box",
    welcome_subtitle: "Broadcast Box er et værktøj, der gør det muligt at streame video i høj kvalitet i realtid ved hjælp af de nyeste video-codecs og WebRTC-teknologi.",

    toggle_watch: "Jeg vil se",
    toggle_stream: "Jeg vil streame",

    stream_input_label: "Stream-nøgle",
    stream_input_placeholder_share: "Indsæt nøglen til det stream, du vil dele",
    stream_input_placeholder_join: "Indsæt nøglen til det stream, du vil deltage i",

    stream_button_stream_start: "Start stream",
    stream_button_stream_join: "Deltag i stream",
  },

  available_streams: {
    title: "Aktuelle streams",
    stream_join_message: "Klik på et stream for at deltage",
    no_streams_message: "Ingen streams tilgængelige i øjeblikket",
  },

  statistics: {
    title: "Statistik",
    no_statistics_available: "Ingen statistik tilgængelig i øjeblikket",
    no_sessions: "Ingen sessioner tilgængelige",

    whep_sessions: "WHEP-sessioner",

    video: "Video",
    video_tracks: "Videospor",
    video_track_not_available: "Ingen videospor",

    audio: "Lyd",
    audio_tracks: "Lydspor",
    audio_track_not_available: "Ingen lydspor",

    button_watch_stream: "Se stream",

    rid: "RID",
    layer: "Lag",
    packets_received: "Modtagne pakker",
    packets_written: "Sendte pakker",
    last_key_frame: "Sidste nøglebillede",
    timestamp: "Tidsstempel",
    sequence_number: "Sekvensnummer"
  },

  player_header: {
    error: "Fejl",
    success: "Succes",
    warning: "Advarsel",

    mediaAccessError_default: "Kunne ikke få adgang til dine medieenheder",
    mediaAccessError_noMediaDevices: "MediaDevices API blev ikke fundet. Udgivelse i Broadcast Box kræver HTTPS",
    mediaAccessError_notAllowedError: "Du kan ikke streame med dit kamera, adgangen er blevet deaktiveret.",
    mediaAccessError_notFoundError: "Det ser ud til, at du ikke har et kamera, eller at adgangen er blokeret\nTjek kameraindstillinger, browser- og systemtilladelser.",

    connection_established: "Live: Streamer i øjeblikket til",
    connection_disconnected: "WebRTC er afbrudt eller kunne ikke oprette forbindelse",
    connection_failed: "Kunne ikke starte Broadcast Box-session",
    connection_has_packetloss: "WebRTC oplever pakketab",

    publish_screen: "Del skærm/vindue/faneblad",
    publish_webcam: "Del webcam",

    button_end_stream: "Afslut stream"
  },

  player_page: {
    cinema_mode_disable: "Deaktiver biotilstand",
    cinema_mode_enable: "Aktiver biotilstand",

    modal_add_stream_title: "Tilføj stream",
    modal_add_stream_message: "Indsæt stream-nøgle for at tilføje til multi-stream",
    modal_add_stream_placeholder: "Indsæt nøglen til den stream, du vil tilføje",

    stream_status_offline: "Offline"
  },

  player: {
    message_is_not_online: "streamer ikke i øjeblikket",
    message_loading_video: "Indlæser video"
  },

  stream_status: {
    message_current_viewers: "Nuværende seere"
  },

  profile_settings: {
    title: "Profilindstillinger",
    subTitle: "Konfigurer streamingprofil",

    toggle_stream_privacy_label: "Privatliv",
    toggle_stream_privacy_title_left: "Privat",
    toggle_stream_privacy_title_right: "Offentlig",

    input_motd_label: "Dagens besked",
    button_save_label: "Gem"
  },

  admin_login: {
    login_input_dialog_title: "Login",
    login_input_dialog_message: "Indsæt admin-token for at logge ind",
    login_input_dialog_placeholder: "Indsæt admin-token for at logge ind",
    error_message_login_failed: "Login mislykkedes",
    button_login_text: "Login"
  },

  admin_page: {
    title: "Adminportal",
    menu_api: "API",
    menu_logging: "Logning",
    menu_logout: "Log ud",
    menu_profiles: "Profiler",
    menu_settings: "Indstillinger",
    menu_status: "Status"
  },

  admin_page_api: {
    title: "API-indstillinger",
    table_header_setting_name: "Indstilling",
    table_header_value: "Værdi"
  },

  admin_page_logging: {
    title: "Logning",
    table_header_setting_name: "Indstilling",
    table_header_value: "Værdi"
  },

  admin_page_profiles: {
    title: "Profiloversigt",

    add_profile_modal_title: "Tilføj profil",
    add_profile_modal_message: "Indsæt en nøgle for at tilføje en ny stream",
    add_profile_modal_placeholder: "Skriv ny stream-nøgle her",

    remove_profile_modal_title: "Fjern profil",
    remove_profile_modal_message: "Er du sikker på, at du vil fjerne",

    table_header_stream_key: "Stream-nøgle",
    table_header_is_public: "Er offentlig",
    table_header_motd: "Motd",
    table_header_token: "Token",

    button_add_profile: "Tilføj profil",

    yes: "Ja",
    no: "Nej"
  },

  admin_page_status_page: {
    title: "Stream-statusoversigt",

    table_header_stream_key: "Stream-nøgle",
    table_header_is_public: "Er offentlig",
    table_header_video_tracks: "Videospor",
    table_header_audio_tracks: "Lydspor",
    table_header_sessions: "Sessioner",
    table_header_total_packets: "Samlede pakker",

    yes: "ja",
    no: "nej"
  },

  shared_component_card: {
    button_accept: "Acceptér"
  },
  shared_component_text_input_modal: {
    button_accept: "Acceptér"
  },
  shared_component_text_input_dialog: {
    button_accept: "Acceptér"
  },
}
export default locale_da
