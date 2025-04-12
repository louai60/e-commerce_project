'use client';

import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext'; 
import { api } from '@/lib/api';
import Breadcrumb from '@/components/Common/Breadcrumb';
import { toast } from 'react-hot-toast';
import SignOutButton from '@/components/Auth/SignOut';

interface UserData {
  user_id: number;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  phone_number: string;
  user_type: string;
  role: string;
  account_status: string;
  created_at: string;
  updated_at: string;
  last_login: string;
}

export default function ProfilePage() {
  const { isAuthenticated, isLoading } = useAuth();
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [userData, setUserData] = useState<UserData | null>(null);

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      const user = JSON.parse(userStr);
      setUserData(user);
    }
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!userData) return;
    
    const { name, value } = e.target;
    setUserData(prev => ({
      ...prev!,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userData) return;

    setLoading(true);
    try {
      const updateData = {
        first_name: userData.first_name,
        last_name: userData.last_name,
        phone_number: userData.phone_number,
      };

      const response = await api.put('/api/v1/users/profile', updateData);
      localStorage.setItem('user', JSON.stringify({ ...userData, ...updateData }));
      toast.success('Profile updated successfully');
      setIsEditing(false);
    } catch (error: any) {
      console.error('Failed to update profile:', error);
      toast.error(error.response?.data?.message || 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  if (isLoading || !userData) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return null;
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <>
      <Breadcrumb title="Profile" pages={["profile"]} />
      
      <section className="overflow-hidden py-20 bg-gray-2">
        <div className="max-w-[1170px] w-full mx-auto px-4 sm:px-8 xl:px-0">
          <div className="bg-white shadow-1 rounded-xl p-4 sm:p-7.5 xl:p-11">
            <div className="flex justify-between items-center mb-8">
              <h2 className="font-semibold text-xl sm:text-2xl xl:text-heading-5 text-dark">
                Profile Information
              </h2>
              <div className="flex gap-4">
                <button
                  onClick={() => setIsEditing(!isEditing)}
                  className="inline-flex items-center justify-center rounded-md bg-primary py-2 px-6 text-white hover:bg-opacity-90"
                >
                  {isEditing ? 'Cancel' : 'Edit Profile'}
                </button>
                <SignOutButton />
              </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                {/* Editable Fields */}
                <div className="w-full">
                  <label htmlFor="first_name" className="block mb-2.5 font-medium">
                    First Name
                  </label>
                  <input
                    type="text"
                    name="first_name"
                    id="first_name"
                    value={userData.first_name}
                    onChange={handleChange}
                    disabled={!isEditing}
                    className="rounded-md border border-gray-3 bg-gray-1 w-full py-2.5 px-5 outline-none"
                  />
                </div>

                <div className="w-full">
                  <label htmlFor="last_name" className="block mb-2.5 font-medium">
                    Last Name
                  </label>
                  <input
                    type="text"
                    name="last_name"
                    id="last_name"
                    value={userData.last_name}
                    onChange={handleChange}
                    disabled={!isEditing}
                    className="rounded-md border border-gray-3 bg-gray-1 w-full py-2.5 px-5 outline-none"
                  />
                </div>

                <div className="w-full">
                  <label htmlFor="phone_number" className="block mb-2.5 font-medium">
                    Phone Number
                  </label>
                  <input
                    type="tel"
                    name="phone_number"
                    id="phone_number"
                    value={userData.phone_number}
                    onChange={handleChange}
                    disabled={!isEditing}
                    className="rounded-md border border-gray-3 bg-gray-1 w-full py-2.5 px-5 outline-none"
                  />
                </div>

                {/* Read-only Fields */}
                <div className="w-full">
                  <label className="block mb-2.5 font-medium">Email</label>
                  <input
                    type="email"
                    value={userData.email}
                    disabled
                    className="rounded-md border border-gray-3 bg-gray-2 w-full py-2.5 px-5"
                  />
                </div>

                <div className="w-full">
                  <label className="block mb-2.5 font-medium">Username</label>
                  <input
                    type="text"
                    value={userData.username}
                    disabled
                    className="rounded-md border border-gray-3 bg-gray-2 w-full py-2.5 px-5"
                  />
                </div>

                <div className="w-full">
                  <label className="block mb-2.5 font-medium">Role</label>
                  <input
                    type="text"
                    value={userData.role}
                    disabled
                    className="rounded-md border border-gray-3 bg-gray-2 w-full py-2.5 px-5"
                  />
                </div>

                <div className="w-full">
                  <label className="block mb-2.5 font-medium">Account Status</label>
                  <input
                    type="text"
                    value={userData.account_status}
                    disabled
                    className="rounded-md border border-gray-3 bg-gray-2 w-full py-2.5 px-5"
                  />
                </div>

                <div className="w-full">
                  <label className="block mb-2.5 font-medium">Last Login</label>
                  <input
                    type="text"
                    value={formatDate(userData.last_login)}
                    disabled
                    className="rounded-md border border-gray-3 bg-gray-2 w-full py-2.5 px-5"
                  />
                </div>

                <div className="w-full">
                  <label className="block mb-2.5 font-medium">Member Since</label>
                  <input
                    type="text"
                    value={formatDate(userData.created_at)}
                    disabled
                    className="rounded-md border border-gray-3 bg-gray-2 w-full py-2.5 px-5"
                  />
                </div>
              </div>

              {isEditing && (
                <div className="flex justify-end mt-8">
                  <button
                    type="submit"
                    disabled={loading}
                    className="inline-flex items-center justify-center rounded-md bg-primary py-3 px-10 text-white hover:bg-opacity-90 disabled:opacity-50"
                  >
                    {loading ? 'Saving...' : 'Save Changes'}
                  </button>
                </div>
              )}
            </form>
          </div>
        </div>
      </section>
    </>
  );
}
