import { useCallback, useContext, useLayoutEffect, useState } from "react";
import Card from "../shared/Card";
import Input from "../shared/Input";
import Toggle from "../shared/Toggle";
import Button from "../shared/Button";
import { LocaleContext } from "../../providers/LocaleProvider";
import { AccountContext } from "../../providers/AccountProvider";
import toBase64Utf8 from "../../utilities/base64";

interface Profile {
  streamKey: string
  isPublic: string
  isActive: boolean
  motd: string
}

interface ProfileSettingsProps {
  stateHasChanged: (isActive: boolean, streamKey: string) => void;
}

export default function ProfileSettings(props: ProfileSettingsProps) {
  const { locale } = useContext(LocaleContext)
  const { account } = useContext(AccountContext)
  const { stateHasChanged } = props
  const streamKey = location.pathname.split('/').pop()
  const authToken = account?.streamKey === streamKey && account.streamToken ? account.streamToken : streamKey

  const [profileType, setProfileType] = useState<"Public" | "Reserved">("Public")
  const [isPublic, setIsPublic] = useState<"Public" | "Private">("Public")
  const [motd, setMotd] = useState<string>("")

  const getSettings = useCallback(() => {
    return fetch(`/api/whip/profile`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(authToken)}`,
      },
    }).then((result) => {
      if (result.status > 400 && result.status < 500 || result.status === 204) {
        return;
      }

      return result.json()
    }).then((result: Profile | undefined) => {

      if (result === undefined) {
        setProfileType(() => "Public")
        return
      }

      setProfileType(() => "Reserved")
      setIsPublic(() => result.isPublic ? "Public" : "Private")
      setMotd(() => result.motd)
      stateHasChanged?.(result.isActive, result.streamKey)
    });
  }, [authToken, stateHasChanged])

  const updateSettings = useCallback(() => {
    fetch(`/api/whip/profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(authToken)}`,
      },
      body: JSON.stringify({
        motd: motd,
        isPublic: isPublic === "Public"
      }),
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        return;
      }

      void getSettings();
    });
  }, [authToken, getSettings, isPublic, motd]);

  useLayoutEffect(() => {
    void getSettings()
  }, [getSettings])

  if (profileType === "Public") {
    return <></>
  }

  return (
    <Card
      title={locale.profile_settings.title}
      subTitle={locale.profile_settings.subTitle}
      classNames="notice-window"
    >
      <div className="form-stack">
        <Input
          label="stream title"
          value={motd}
          setValue={setMotd}
          placeholder="enter a title for your stream"
        />

        <Toggle
          label={locale.profile_settings.toggle_stream_privacy_label}
          titleLeft={locale.profile_settings.toggle_stream_privacy_title_left}
          onClickLeft={() => setIsPublic(() => "Private")}

          titleRight={locale.profile_settings.toggle_stream_privacy_title_right}
          onClickRight={() => setIsPublic(() => "Public")}

          selected={isPublic === "Public" ? "Right" : "Left"}
        />

        <Button
          title={locale.profile_settings.button_save_label}
          color='Accept'
          onClick={updateSettings}
        />
      </div>
    </Card>
  )
}
