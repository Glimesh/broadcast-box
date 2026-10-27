import { useEffect, useRef, useState } from "react";
import Button from "../../shared/Button";
import Input from "../../shared/Input";
import { clearAdminToken, getAdminAuthorizationHeader, isInvalidAdminSessionResponse } from "../adminAuth";

interface Emote {
  name: string;
  file: string;
  url: string;
  animated: boolean;
}

const ALLOWED_EXTENSIONS = [".png", ".gif", ".jpg", ".jpeg", ".webp", ".apng"];

const sanitizeEmoteName = (value: string) => {
  return value
    .replace(/\.[^/.]+$/, "")
    .replace(/[^a-zA-Z0-9_-]/g, "_")
    .replace(/^_+|_+$/g, "")
    .slice(0, 32);
};

const EmotesPage = () => {
  const [emotes, setEmotes] = useState<Emote[]>([]);
  const [name, setName] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const [errorMessage, setErrorMessage] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState<{ done: number; total: number } | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const refreshEmotes = () => {
    fetch(`/api/admin/emotes`, {
      method: "GET",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }

        return result.json();
      })
      .then((result: Emote[]) => {
        if (Array.isArray(result)) {
          setEmotes(result);
        }
      })
      .catch(() => setErrorMessage("Could not load emotes"));
  };

  useEffect(() => {
    refreshEmotes();
  }, []);

  const addFiles = (incoming: FileList | File[]) => {
    const nextFiles = Array.from(incoming).filter((file) => {
      const lowerName = file.name.toLowerCase();
      return ALLOWED_EXTENSIONS.some((extension) => lowerName.endsWith(extension));
    });

    if (nextFiles.length === 0) {
      setErrorMessage("Only PNG, GIF, JPG, WEBP, or APNG files are supported");
      return;
    }

    setErrorMessage("");
    setFiles((current) => {
      const existing = new Set(current.map((file) => `${file.name}_${file.size}`));
      return [
        ...current,
        ...nextFiles.filter((file) => !existing.has(`${file.name}_${file.size}`)),
      ];
    });
  };

  const removeFile = (index: number) => {
    setFiles((current) => current.filter((_, fileIndex) => fileIndex !== index));
  };

  const uploadEmotes = async () => {
    if (files.length === 0) {
      setErrorMessage("Choose one or more image or GIF files first");
      return;
    }

    setErrorMessage("");
    setIsUploading(true);
    setProgress({ done: 0, total: files.length });

    const failures: string[] = [];
    const useSingleName = files.length === 1 && name.trim() !== "";

    for (let index = 0; index < files.length; index += 1) {
      const file = files[index];
      const emoteName = useSingleName ? name.trim() : sanitizeEmoteName(file.name);

      const formData = new FormData();
      formData.append("name", emoteName);
      formData.append("file", file);

      try {
        const result = await fetch(`/api/admin/emotes`, {
          method: "POST",
          headers: {
            Authorization: getAdminAuthorizationHeader(),
          },
          body: formData,
        });

        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }

        if (!result.ok) {
          failures.push(`${file.name}: ${await result.text()}`);
        }
      } catch {
        failures.push(`${file.name}: upload failed`);
      }

      setProgress({ done: index + 1, total: files.length });
    }

    setIsUploading(false);
    setProgress(null);
    setName("");
    setFiles([]);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }

    if (failures.length > 0) {
      setErrorMessage(failures.join("\n"));
    }

    refreshEmotes();
  };

  const deleteEmote = async (emoteName: string) => {
    const result = await fetch(`/api/admin/emotes?name=${encodeURIComponent(emoteName)}`, {
      method: "DELETE",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
    });

    if (isInvalidAdminSessionResponse(result.status)) {
      clearAdminToken();
      return;
    }

    refreshEmotes();
  };

  return (
    <div className="flex w-full flex-col gap-6">
      <div>
        <h1 className="panel-heading text-base">custom emotes</h1>
        <p className="panel-copy">
          upload static or animated emotes. people can use them in chat by typing <code className="emote-code">:name:</code> or picking them from the emote menu.
        </p>
      </div>

      <div className="pixel-panel">
        <h2 className="panel-heading mb-3">add emotes</h2>
        <div className="grid grid-cols-1 md:grid-cols-[1fr_1fr_auto] gap-3 items-end">
          <Input
            label="emote name (single file only)"
            value={name}
            setValue={setName}
            placeholder="e.g. poggers"
            classNames="mb-0"
          />
          <div>
            <label className="pixel-label">images / GIFs / WebP / APNG</label>
            <input
              ref={fileInputRef}
              type="file"
              multiple
              accept="image/png,image/gif,image/jpeg,image/webp,image/apng"
              onChange={(event) => {
                if (event.target.files) {
                  addFiles(event.target.files);
                }
              }}
              className="file-input w-full text-sm"
            />
          </div>
          <Button
            title={isUploading ? `Uploading ${progress?.done ?? 0}/${progress?.total ?? files.length}...` : `Upload ${files.length || ""} emote${files.length === 1 ? "" : "s"}`}
            color="Accept"
            onClick={uploadEmotes}
            isDisabled={isUploading || files.length === 0}
          />
        </div>

        <div
          onDragOver={(event) => {
            event.preventDefault();
            setIsDragging(true);
          }}
          onDragLeave={() => setIsDragging(false)}
          onDrop={(event) => {
            event.preventDefault();
            setIsDragging(false);
            if (event.dataTransfer.files) {
              addFiles(event.dataTransfer.files);
            }
          }}
          className={`file-drop mt-3 ${isDragging ? "file-drop--active" : ""}`}
        >
          drag and drop one or many emote files here ♡
        </div>

        {files.length > 0 && (
          <div className="mt-3">
            <div className="pixel-label">
              selected ({files.length}){files.length > 1 ? " — names will come from filenames" : ""}
            </div>
            <div className="flex flex-wrap gap-2">
              {files.map((file, index) => (
                <div key={`${file.name}_${file.size}_${index}`} className="file-chip flex items-center gap-2 px-2 py-1 text-xs">
                  <span className="max-w-48 truncate">{file.name}</span>
                  <button
                    type="button"
                    onClick={() => removeFile(index)}
                    className="file-chip__remove"
                    title="remove"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {progress && (
          <div className="mt-3">
            <div className="progress-track">
              <div
                className="progress-fill"
                style={{ width: `${(progress.done / Math.max(progress.total, 1)) * 100}%` }}
              />
            </div>
          </div>
        )}

        {errorMessage && <p className="form-error mt-2 whitespace-pre-line">{errorMessage}</p>}
      </div>

      <div>
        <h2 className="panel-heading mb-3">uploaded emotes ({emotes.length})</h2>
        {emotes.length === 0 ? (
          <p className="muted text-sm">no custom emotes yet ♡</p>
        ) : (
          <div className="emote-grid">
            {emotes.map((emote) => (
              <div key={emote.name} className="emote-card group">
                <div className="emote-card__preview">
                  <img src={emote.url} alt={emote.name} />
                </div>
                <div className="text-center w-full">
                  <div className="w-full truncate text-center text-xs font-semibold">:{emote.name}:</div>
                  <div className="text-[0.65rem] muted">{emote.animated ? "animated" : "static"}</div>
                </div>
                <button
                  onClick={() => deleteEmote(emote.name)}
                  className="pixel-btn pixel-btn--danger w-full"
                >
                  delete
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default EmotesPage;