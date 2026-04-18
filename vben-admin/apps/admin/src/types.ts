export interface MenuItem {
  id: number;
  parent_id: number;
  name: string;
  title: string;
  path: string;
  component: string;
  menu_type: string;
  icon: string;
  sort: number;
  permission_code: string;
  children?: MenuItem[];
}

export interface AdminUser {
  id: number;
  username: string;
  nickname: string;
  is_super: boolean;
  role_ids: number[];
  access_codes: string[];
}
