import { useContext, useState } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";

interface ApiSettingsResult {
  name: string
  value: string
}

const ApiPage = () => {
  const { locale } = useContext(LocaleContext)
  const [response] = useState<ApiSettingsResult[]>()

  return (
    <div className="flex w-full flex-col gap-4">
      <h1 className="panel-heading text-base">{locale.admin_page_api.title}</h1>

      <div className="admin-table-wrap">
        <table className="data-table">
          <thead>
            <tr>
              <th>{locale.admin_page_api.table_header_setting_name}</th>
              <th>{locale.admin_page_api.table_header_value}</th>
            </tr>
          </thead>
          <tbody>
            {response?.map((setting, index) => {
              return (
                <tr key={index} className="border-t">
                  <td>{setting.name}</td>
                  <td>{setting.value}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
export default ApiPage;
