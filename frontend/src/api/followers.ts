import { apiURL } from "./api";

export type ConnectionUser = {
  id: string;
  firstName: string;
  lastName: string;
  nickname?: string;
  avatar?: string;
};

export const getUserFollowers = async (userID: string): Promise<ConnectionUser[]> => {
  const response = await fetch(apiURL(`/api/users/${userID}/followers`), {
    method: 'GET',
    credentials: 'include',
  });
  
  if (!response.ok) throw new Error('Failed to fetch followers');
  
  return await response.json();
}