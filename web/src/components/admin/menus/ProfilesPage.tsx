import { ArrowPathIcon, XMarkIcon } from "@heroicons/react/16/solid";
import React, { useEffect, useState } from "react";
import Button from "../../shared/Button";
import ModalTextInput from "../../shared/ModalTextInput";
import ModalMessageBox from "../../shared/ModalMessageBox";

const ADMIN_TOKEN = "adminToken";

interface Profile {
  streamKey: string
  token: string
  isPublic: boolean
  motd: string
}

const ProfilesPage = () => {
  const [response, setResponse] = useState<Profile[]>()
  const [isAddProfileModalOpen, setIsAddProfileModalOpen] = useState<boolean>(false)
  const [isRemoveProfileModalOpen, setIsRemoveProfileModalOpen] = useState<string>("")
  const [errorMessage, setErrorMessage] = useState<string>()

  const refreshProfiles = () => {
    fetch(`/api/admin/profiles`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        localStorage.removeItem(ADMIN_TOKEN)
        return;
      }

      return result.json();
    }).then((result) => {
      setResponse(() => result)
    });
  };
  const resetProfileToken = (streamKey: string) => {
    fetch(`/api/admin/profiles/reset-token`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
      body: JSON.stringify({ streamKey: streamKey })
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        localStorage.removeItem(ADMIN_TOKEN)
        return;
      }

      refreshProfiles()
    });
  };
  const addProfile = (streamKey: string) => {
    fetch(`/api/admin/profiles/add-profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
      body: JSON.stringify({ streamKey: streamKey })
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        localStorage.removeItem(ADMIN_TOKEN)
        return;
      }

      if (result.status === 400) {
        result.text()
          .then((resultText) => setErrorMessage(() => resultText))

        return
      }

      setIsAddProfileModalOpen(() => false)
      refreshProfiles()
    });
  };
  const removeProfile = (streamKey: string) => {
    fetch(`/api/admin/profiles/remove-profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
      body: JSON.stringify({ streamKey: streamKey })
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        localStorage.removeItem(ADMIN_TOKEN)
        return;
      }

      if (result.status === 400) {
        result.text()
          .then((resultText) => setErrorMessage(() => resultText))

        return
      }

      setIsRemoveProfileModalOpen(() => "")
      refreshProfiles()
    });
  };

  useEffect(() => {
    refreshProfiles()
  }, [])

  return (
    <div className="p-6 w-full h-full max-w-6xl mx-auto flex flex-col justify-between">
      <h1 className="text-3xl font-bold mb-6">Profiles Overview</h1>

      <div className="overflow-x-auto h-full">
        <ModalTextInput
          title="Add Profile"
          message="Insert a key to add a new stream"
          errorMessage={errorMessage}
          placeholder="Write new stream key here"
          isOpen={isAddProfileModalOpen}
          canCloseOnBackgroundClick={true}
          onAccept={(result) => addProfile(result.toString())}
          onDeny={() => setIsAddProfileModalOpen(false)}
        />

        <ModalMessageBox
          title="Remove profile"
          message={`Are you sure you would like to remove ${isRemoveProfileModalOpen}`}
          errorMessage={errorMessage}
          isOpen={isRemoveProfileModalOpen !== ""}
          canCloseOnBackgroundClick={true}
          onAccept={() => removeProfile(isRemoveProfileModalOpen)}
          onDeny={() => setIsRemoveProfileModalOpen("")}
        />

        <table className="min-w-full rounded-lg">
          <thead className=" text-white">
            <tr>
              <th className="px-4 py-2 text-left">Stream Key</th>
              <th className="px-4 py-2 text-left">Is Public</th>
              <th className="px-4 py-2 text-left">Motd</th>
              <th className="px-4 py-2 text-left">Token</th>
              <th className="px-4 py-2 text-left"></th>
            </tr>
          </thead>
          <tbody>
            {response?.map((profile, index) => {
              return (
                <tr key={index} className="border-t">
                  <td className="px-4 py-2 font-medium ">{profile.streamKey}</td>
                  <td className="px-4 py-2 font-medium ">{profile.isPublic ? "Yes" : "No"}</td>
                  <td className="px-4 py-2">{profile.motd}</td>
                  <td className="px-4 py-2 flex flex-row justify-between" title="To be implemented">
                    {profile.token}

                    <ArrowPathIcon
                      className="ml-2 h-6"
                      onClick={() => resetProfileToken(profile.streamKey)} />
                  </td>
                  <td className="px-4 py-2">
                    <XMarkIcon
                      className="ml-2 h-6 text-red-700"
                      onClick={() => setIsRemoveProfileModalOpen(() => profile.streamKey)}
                    />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <Button
        title="Add profile"
        onClick={() => setIsAddProfileModalOpen(() => true)}
      />
    </div>
  );
}
export default ProfilesPage;
