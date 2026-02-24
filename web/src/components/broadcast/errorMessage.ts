import { localeInterface } from "../../locale/localeInterface";

export enum ErrorMessageEnum {
  NoMediaDevices,
  NotAllowedError,
  NotFoundError
}

export function getMediaErrorMessage(locale: localeInterface, value: ErrorMessageEnum): string {
  switch (value) {
    case ErrorMessageEnum.NoMediaDevices:
      return locale.player_header.mediaAccessError_noMediaDevices;
    case ErrorMessageEnum.NotFoundError:
      return locale.player_header.mediaAccessError_notFoundError;
    case ErrorMessageEnum.NotAllowedError:
      return locale.player_header.mediaAccessError_notAllowedError;
    default:
      return locale.player_header.mediaAccessError_default;
  }
}
