import { ArrowPathIcon } from "@heroicons/react/16/solid";
import React, { useEffect, useState } from "react";

const ADMIN_TOKEN = "adminToken";

interface Profile {
  streamKey: string
  token: string
  isPublic: boolean
  motd: string
}

const ProfilesPage = () => {
  const [response, setResponse] = useState<Profile[]>()

  const refreshProfiles = () => {
    fetch(`/api/admin/profiles`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
    })
      .then((result) => {
        if (result.status >= 400 && result.status < 500) {
          localStorage.removeItem(ADMIN_TOKEN)
          return;
        }

        return result.json();
      })
      .then((result) => {
        setResponse(() => result)
      });
  };
  const resetToken = (streamKey: string) => {
    fetch(`/api/admin/profiles/reset-token`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
      body: JSON.stringify({ streamKey: streamKey })
    })
      .then((result) => {
        if (result.status >= 400 && result.status < 500) {
          localStorage.removeItem(ADMIN_TOKEN)
          return;
        }

        refreshProfiles()
      });
  };

  useEffect(() => {
    refreshProfiles()
  }, [])

  return (
    <div className="p-6 w-full max-w-6xl mx-auto">
      <h1 className="text-3xl font-bold mb-6">Profiles Overview</h1>

      <div className="overflow-x-auto">
        <table className="min-w-full rounded-lg">
          <thead className=" text-white">
            <tr>
              <th className="px-4 py-2 text-left">Stream Key</th>
              <th className="px-4 py-2 text-left">Is Public</th>
              <th className="px-4 py-2 text-left">Motd</th>
              <th className="px-4 py-2 text-left">Token</th>
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

                    <ArrowPathIcon onClick={() => resetToken(profile.streamKey)} className="ml-2 h-6" />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
export default ProfilesPage;

