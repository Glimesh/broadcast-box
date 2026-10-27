import { useEffect, useState } from "react";
import Button from "../../shared/Button";
import Input from "../../shared/Input";
import {
  clearAdminToken,
  getAdminAuthorizationHeader,
  isInvalidAdminSessionResponse,
} from "../adminAuth";

interface MusicPlaylist {
  title: string;
  subtitle: string;
  videoIds: string[];
  defaultVolume: number;
}

const MusicPage = () => {
  const [title, setTitle] = useState("");
  const [subtitle, setSubtitle] = useState("");
  const [videoIds, setVideoIds] = useState("");
  const [defaultVolume, setDefaultVolume] = useState(30);
  const [message, setMessage] = useState("");

  const refresh = () => {
    fetch("/api/admin/music", {
      method: "GET",
      headers: { Authorization: getAdminAuthorizationHeader() },
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }
        return result.json();
      })
      .then((result: MusicPlaylist | undefined) => {
        if (result) {
          setTitle(result.title);
          setSubtitle(result.subtitle);
          setVideoIds(result.videoIds.join("\n"));
          setDefaultVolume(result.defaultVolume);
        }
      })
      .catch(() => setMessage("Could not load music playlist"));
  };

  useEffect(() => {
    refresh();
  }, []);

  const save = () => {
    const ids = videoIds
      .split(/\n|,/)
      .map((value) => value.trim())
      .filter(Boolean);

    fetch("/api/admin/music", {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        title,
        subtitle,
        videoIds: ids,
        defaultVolume,
      }),
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }
        if (!result.ok) {
          return result.text().then((text) => setMessage(text));
        }
        setMessage("music playlist saved ♡");
        refresh();
      })
      .catch(() => setMessage("Could not save music playlist"));
  };

  return (
    <div className="flex w-full flex-col gap-4">
      <h1 className="panel-heading text-base">music playlist</h1>
      <p className="panel-copy">changes apply to the site music player.</p>

      <div className="form-stack">
        <Input label="title" value={title} setValue={setTitle} placeholder="frutiger aero radio" />
        <Input label="subtitle" value={subtitle} setValue={setSubtitle} placeholder="nintendo mix · eShop / Wii / 3DS / DS" />
        <label className="form-field">
          <span className="pixel-label">video ids (one per line)</span>
          <textarea
            value={videoIds}
            onChange={(event) => setVideoIds(event.target.value)}
            className="pixel-textarea"
            rows={8}
          />
        </label>
        <Input
          label="default volume"
          value={String(defaultVolume)}
          setValue={(value) => setDefaultVolume(Number(value) || 0)}
          placeholder="30"
        />
        <Button title="save playlist ♡" color="Accept" onClick={save} />
        {message && <p className="form-error mt-2">{message}</p>}
      </div>
    </div>
  );
};

export default MusicPage;
