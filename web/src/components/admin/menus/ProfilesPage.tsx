import { ArrowPathIcon, XMarkIcon } from "@heroicons/react/16/solid";
import React, { useContext, useEffect, useState } from "react";
import Button from "../../shared/Button";
import ModalTextInput from "../../shared/ModalTextInput";
import ModalMessageBox from "../../shared/ModalMessageBox";
import { LocaleContext } from "../../../providers/LocaleProvider";
import { getIcon } from "../../shared/Icons";
import toBase64Utf8 from "../../../utilities/base64";

const ADMIN_TOKEN = "adminToken";

interface Profile {
  streamKey: string;
  token: string;
  isPublic: boolean;
  motd: string;
}

const ProfilesPage = () => {
  const { locale } = useContext(LocaleContext);
  const [response, setResponse] = useState<Profile[]>();
  const [isAddProfileModalOpen, setIsAddProfileModalOpen] = useState<boolean>(false);
  const [isRemoveProfileModalOpen, setIsRemoveProfileModalOpen] = useState<string>("");
  const [errorMessage, setErrorMessage] = useState<string>();

  const copyTokenToClipboard = (token: string) => navigator.clipboard.writeText(token)

  const refreshProfiles = () => {
    fetch(`/api/admin/profiles`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(localStorage.getItem(ADMIN_TOKEN))}`,
      },
    })
      .then((result) => {
        if (result.status > 400 && result.status < 500) {
          localStorage.removeItem(ADMIN_TOKEN);
          return;
        }

        return result.json();
      })
      .then((result) => {
        setResponse(() => result);
      });
  };
  const resetProfileToken = (streamKey: string) => {
    fetch(`/api/admin/profiles/reset-token`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(localStorage.getItem(ADMIN_TOKEN))}`,
      },
      body: JSON.stringify({ streamKey: streamKey }),
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        localStorage.removeItem(ADMIN_TOKEN);
        return;
      }

      refreshProfiles();
    });
  };
  const addProfile = (streamKey: string) => {
    fetch(`/api/admin/profiles/add-profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(localStorage.getItem(ADMIN_TOKEN))}`,
      },
      body: JSON.stringify({ streamKey: streamKey }),
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        localStorage.removeItem(ADMIN_TOKEN);
        return;
      }

      if (result.status === 400) {
        result.text().then((resultText) => setErrorMessage(() => resultText));

        return;
      }

      setIsAddProfileModalOpen(() => false);
      refreshProfiles();
    });
  };
  const removeProfile = (streamKey: string) => {
    fetch(`/api/admin/profiles/remove-profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(localStorage.getItem(ADMIN_TOKEN))}`,
      },
      body: JSON.stringify({ streamKey: streamKey }),
    }).then((result) => {
      if (result.status > 400 && result.status < 500) {
        localStorage.removeItem(ADMIN_TOKEN);
        return;
      }

      if (result.status === 400) {
        result.text().then((resultText) => setErrorMessage(() => resultText));

        return;
      }

      setIsRemoveProfileModalOpen(() => "");
      refreshProfiles();
    });
  };

  useEffect(() => {
    refreshProfiles();
  }, []);

  return (
    <div className="p-6 w-full h-full max-w-6xl mx-auto flex flex-col justify-between">
      <h1 className="text-3xl font-bold mb-6">{locale.admin_page_profiles.title}</h1>

      <div className="overflow-x-auto h-full">
        <ModalTextInput
          title={locale.admin_page_profiles.add_profile_modal_title}
          message={locale.admin_page_profiles.add_profile_modal_message}
          errorMessage={errorMessage}
          placeholder={locale.admin_page_profiles.add_profile_modal_placeholder}
          isOpen={isAddProfileModalOpen}
          canCloseOnBackgroundClick={true}
          onAccept={(result) => addProfile(result.toString())}
          onDeny={() => setIsAddProfileModalOpen(false)}
        />

        <ModalMessageBox
          title={locale.admin_page_profiles.remove_profile_modal_title}
          message={locale.admin_page_profiles.remove_profile_modal_message + " " + isRemoveProfileModalOpen}
          errorMessage={errorMessage}
          isOpen={isRemoveProfileModalOpen !== ""}
          canCloseOnBackgroundClick={true}
          onAccept={() => removeProfile(isRemoveProfileModalOpen)}
          onDeny={() => setIsRemoveProfileModalOpen("")}
        />

        <table className="min-w-full rounded-lg">
          <thead className=" text-white">
            <tr>
              <th className="px-4 py-2 text-left">{locale.admin_page_profiles.table_header_stream_key}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_profiles.table_header_is_public}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_profiles.table_header_motd}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_profiles.table_header_token}</th>
              <th className="px-4 py-2 text-left"></th>
            </tr>
          </thead>
          <tbody>
            {response?.map((profile, index) => {
              return (
                <tr key={index} className="border-t">
                  <td className="px-4 py-2 font-medium ">
                    {profile.streamKey}
                  </td>
                  <td className="px-4 py-2 font-medium ">
                    {profile.isPublic
                      ? locale.admin_page_profiles.yes
                      : locale.admin_page_profiles.no}
                  </td>
                  <td className="px-4 py-2">{profile.motd}</td>
                  <td className="px-4 py-2 flex flex-row justify-between items-center">
                    <div
                      title="Copy to clipboard"
                      className="flex flex-row gap-1 cursor-pointer"
                      onClick={() => copyTokenToClipboard(profile.token)} >
                      {getIcon("Copy")}
                      {profile.token}
                    </div>

                    <ArrowPathIcon
                      className="ml-2 h-6"
                      onClick={() => resetProfileToken(profile.streamKey)}
                    />
                  </td>
                  <td className="px-4 py-2 items-center">
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
        title={locale.admin_page_profiles.button_add_profile}
        onClick={() => setIsAddProfileModalOpen(() => true)}
      />
    </div>
  );
};
export default ProfilesPage;
