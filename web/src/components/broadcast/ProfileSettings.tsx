import React, { useLayoutEffect, useState } from "react";
import Card from "../shared/Card";
import Input from "../shared/Input";
import Toggle from "../shared/Toggle";
import Button from "../shared/Button";

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
  const streamKey = location.pathname.split('/').pop()

  const [profileType, setProfileType] = useState<"Public" | "Reserved">("Public")
  const [isPublic, setIsPublic] = useState<"Public" | "Private">("Public")
  const [motd, setMotd] = useState<string>("")

  const updateSettings = () => {
    fetch(`/api/whip/profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${streamKey}`,
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
        Authorization: `Bearer ${streamKey}`,
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
  }, [])

  if (profileType === "Public") {
    return <></>
  }

  return (
    <Card
      title="Profile settings"
      subTitle='Configure streaming profile'
    >

      <Input
        label="Message of the day"
        value={motd}
        setValue={setMotd}
      />

      <Toggle
        label="Stream privacy"
        titleLeft='Is Private'
        onClickLeft={() => setIsPublic(() => "Private")}

        titleRight='Is Public'
        onClickRight={() => setIsPublic(() => "Public")}

        selected={isPublic === "Public" ? "Right" : "Left"}
      />

      <Button
        title="Save"
        color='Accept'
        onClick={updateSettings}
      />
    </Card>
  )
}
