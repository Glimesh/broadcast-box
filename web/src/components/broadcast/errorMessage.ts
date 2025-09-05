export enum ErrorMessageEnum {
  NoMediaDevices,
  NotAllowedError,
  NotFoundError
}

export function getMediaErrorMessage(value: ErrorMessageEnum): string {
  switch (value) {
    case ErrorMessageEnum.NoMediaDevices:
      return `MediaDevices API was not found. Publishing in Broadcast Box requires HTTPS 👮`;
    case ErrorMessageEnum.NotFoundError:
      return `Seems like you don't have camera. Or the access to it is blocked\nCheck camera settings, browser permissions and system permissions.`;
    case ErrorMessageEnum.NotAllowedError:
      return `You can't publish stream using your camera, access has been disabled.`;
    default:
      return "Could not access your media device";
  }
}
