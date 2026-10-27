import { ArrowPathIcon, XMarkIcon } from "@heroicons/react/16/solid";
import { useContext, useEffect, useState } from "react";
import Button from "../../shared/Button";
import ModalTextInput from "../../shared/ModalTextInput";
import ModalMessageBox from "../../shared/ModalMessageBox";
import { LocaleContext } from "../../../providers/LocaleProvider";
import { getIcon } from "../../shared/Icons";
import {
  clearAdminToken,
  getAdminAuthorizationHeader,
  isInvalidAdminSessionResponse,
} from "../adminAuth";

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
      .then((result) => {
        setResponse(() => result);
      });
  };
  const resetProfileToken = (streamKey: string) => {
    fetch(`/api/admin/profiles/reset-token`, {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
      body: JSON.stringify({ streamKey: streamKey }),
    }).then((result) => {
      if (isInvalidAdminSessionResponse(result.status)) {
        clearAdminToken();
        return;
      }

      refreshProfiles();
    });
  };
  const addProfile = (streamKey: string) => {
    fetch(`/api/admin/profiles/add-profile`, {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
      body: JSON.stringify({ streamKey: streamKey }),
    }).then((result) => {
      if (isInvalidAdminSessionResponse(result.status)) {
        clearAdminToken();
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
        Authorization: getAdminAuthorizationHeader(),
      },
      body: JSON.stringify({ streamKey: streamKey }),
    }).then((result) => {
      if (isInvalidAdminSessionResponse(result.status)) {
        clearAdminToken();
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
    <div className="flex w-full flex-col gap-4">
      <h1 className="panel-heading text-base">{locale.admin_page_profiles.title}</h1>

      <div className="admin-table-wrap h-full">
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

        <table className="data-table">
          <thead>
            <tr>
              <th>{locale.admin_page_profiles.table_header_stream_key}</th>
              <th>{locale.admin_page_profiles.table_header_is_public}</th>
              <th>{locale.admin_page_profiles.table_header_motd}</th>
              <th>{locale.admin_page_profiles.table_header_token}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {response?.map((profile, index) => {
              return (
                <tr key={index} className="border-t">
                  <td>
                    {profile.streamKey}
                  </td>
                  <td>
                    {profile.isPublic
                      ? locale.admin_page_profiles.yes
                      : locale.admin_page_profiles.no}
                  </td>
                  <td>{profile.motd}</td>
                  <td className="flex flex-row items-center justify-between gap-2">
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
                  <td>
                    <XMarkIcon
                      className="admin-delete ml-2 h-6"
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
