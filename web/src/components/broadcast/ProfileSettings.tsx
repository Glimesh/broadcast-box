import React, { useContext, useLayoutEffect, useState } from "react";
import Card from "../shared/Card";
import Input from "../shared/Input";
import Toggle from "../shared/Toggle";
import Button from "../shared/Button";
import { LocaleContext } from "../../providers/LocaleProvider";
import toBase64Utf8 from "../../utilities/base64";

interface Profile {
  streamKey: string
  isPublic: string
  isActive: boolean
  motd: string
}

interface ProfileSettingsProps {
  // eslint-disable-next-line no-unused-vars
  stateHasChanged: (isActive: boolean, streamKey: string) => void;
}

export default function ProfileSettings(props: ProfileSettingsProps) {
  const { locale } = useContext(LocaleContext)
  const streamKey = location.pathname.split('/').pop()

  const [profileType, setProfileType] = useState<"Public" | "Reserved">("Public")
  const [isPublic, setIsPublic] = useState<"Public" | "Private">("Public")
  const [motd, setMotd] = useState<string>("")

  const updateSettings = () => {
    fetch(`/api/whip/profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(streamKey)}`,
      },
      body: JSON.stringify({
        motd: motd,
        isPublic: isPublic === "Public"
      }),
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        return;
      }

      getSettings();
    });
  };
  const getSettings = () => {
    fetch(`/api/whip/profile`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(streamKey)}`,
      },
    }).then((result) => {
      if (result.status > 400 && result.status < 500 || result.status === 204) {
        return;
      }

      return result.json()
    }).then((result: Profile) => {

      if (result === undefined) {
        setProfileType(() => "Public")
        return
      }

      setProfileType(() => "Reserved")
      setIsPublic(() => result.isPublic ? "Public" : "Private")
      setMotd(() => result.motd)
      props.stateHasChanged?.(result.isActive, result.streamKey)
    });
  };

  useLayoutEffect(() => {
    getSettings()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (profileType === "Public") {
    return <></>
  }

  return (
    <Card
      title={locale.profile_settings.title}
      subTitle={locale.profile_settings.subTitle}
    >

      <Input
        label={locale.profile_settings.input_motd_label}
        value={motd}
        setValue={setMotd}
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
    </Card>
  )
}
