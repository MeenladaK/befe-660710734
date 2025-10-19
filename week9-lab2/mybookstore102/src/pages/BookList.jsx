import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { BookOpenIcon, LogoutIcon } from "@heroicons/react/outline";
import LoadingSpinner from "../components/LoadingSpinner";

const BookList = () => {
  const [books, setBooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const raw = process.env.REACT_APP_API_URL || "http://localhost:8080";
  const apiBase = raw.replace(/\/+$/, "");

  // ใช้ useCallback เพื่อกัน ESLint เตือน และไม่สร้างฟังก์ชันใหม่ทุก render
  const fetchBooks = useCallback(async () => {
    try {
      setLoading(true);
      const res = await fetch(`${apiBase}/api/v1/books`);
      if (!res.ok) throw new Error("ไม่สามารถดึงข้อมูลได้");
      const data = await res.json();
      setBooks(data);
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error("Error fetching books:", err);
    } finally {
      setLoading(false);
    }
  }, [apiBase]);

  useEffect(() => {
    const isAuthenticated = localStorage.getItem("isAdminAuthenticated");
    if (!isAuthenticated) navigate("/login");
  }, [navigate]);

  useEffect(() => {
    fetchBooks();
  }, [fetchBooks]);

  const handleAddBook = () => navigate("/store-manager/add-book");
  const handleEdit = (id) => navigate(`/store-manager/edit-book/${id}`);

  const handleDelete = async (book) => {
    if (window.confirm(`คุณต้องการลบ "${book.title}" ใช่หรือไม่?`)) {
      try {
        const res = await fetch(`${apiBase}/api/v1/books/${book.id}`, {
          method: "DELETE",
        });
        if (!res.ok) throw new Error("ไม่สามารถลบหนังสือได้");
        setBooks((prev) => prev.filter((b) => b.id !== book.id));
        alert("ลบหนังสือเรียบร้อยแล้ว!");
      } catch (err) {
        alert("เกิดข้อผิดพลาด: " + err.message);
      }
    }
  };

  const handleLogout = () => {
    localStorage.removeItem("isAdminAuthenticated");
    navigate("/login");
  };

  if (loading) return <LoadingSpinner />;

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <p className="text-red-600 mb-4">❌ เกิดข้อผิดพลาด: {error}</p>
          <button
            onClick={fetchBooks}
            className="px-6 py-2 bg-viridian-600 text-white rounded-lg hover:bg-viridian-700"
          >
            ลองอีกครั้ง
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-gradient-to-r from-viridian-600 to-green-700 text-white shadow-lg">
        <div className="container mx-auto px-4 py-6 flex justify-between items-center">
          <div className="flex items-center space-x-3">
            <BookOpenIcon className="h-8 w-8" />
            <h1 className="text-2xl font-bold">BookStore - BackOffice</h1>
          </div>
          <button
            onClick={handleLogout}
            className="flex items-center space-x-2 px-4 py-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors"
          >
            <LogoutIcon className="h-5 w-5" />
            <span>ออกจากระบบ</span>
          </button>
        </div>
      </header>

      {/* Content */}
      <div className="container mx-auto px-4 py-8">
        <div className="bg-white rounded-xl shadow-lg p-8">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-2xl font-bold text-gray-900">จัดการหนังสือทั้งหมด</h2>

            <div className="flex gap-2">
              <button
                onClick={handleAddBook}
                className="flex items-center space-x-2 px-5 py-2 bg-green-200 text-viridian-700 rounded-lg font-semibold hover:bg-green-300 transition-colors"
              >
                <span>➕ เพิ่มหนังสือใหม่</span>
              </button>

              {/* ปุ่มแก้ไขหนังสือ */}
              <button
                type="button"
                className="flex items-center space-x-2 px-5 py-2 bg-green-200 text-viridian-700 rounded-lg font-semibold hover:bg-green-300 transition-colors"
              >
                <span>✏️ แก้ไขหนังสือ</span>
              </button>
            </div>
          </div>

          <div className="mb-4 text-gray-600">
            จำนวนหนังสือทั้งหมด{" "}
            <span className="text-viridian-600 font-bold text-xl">{books.length}</span>{" "}
            เล่ม
          </div>

          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead className="bg-gradient-to-r from-viridian-600 to-green-600 text-white">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-semibold">#</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold">ชื่อหนังสือ</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold">ผู้แต่ง</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold">ISBN</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold">ปีที่พิมพ์</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold">ราคา (฿)</th>
                  <th className="px-4 py-3 text-center text-sm font-semibold">จัดการ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {books.length === 0 ? (
                  <tr>
                    <td colSpan="7" className="px-6 py-12 text-center text-gray-500">
                      ไม่พบข้อมูลหนังสือ
                    </td>
                  </tr>
                ) : (
                  books.map((book, index) => (
                    <tr key={book.id} className="hover:bg-green-50 transition-colors">
                      <td className="px-4 py-3 text-gray-900">{index + 1}</td>
                      <td className="px-4 py-3 text-gray-900 font-semibold">{book.title}</td>
                      <td className="px-4 py-3 text-gray-700">{book.author}</td>
                      <td className="px-4 py-3 text-gray-700">{book.isbn}</td>
                      <td className="px-4 py-3 text-gray-700">{book.year}</td>
                      <td className="px-4 py-3 text-green-600 font-semibold">
                        ฿{Number(book.price).toFixed(2)}
                      </td>
                      <td className="px-4 py-3 text-center">
                        <div className="flex justify-center gap-2">
                          <button
                            onClick={() => handleEdit(book.id)}
                            className="px-3 py-2 bg-viridian-600 text-white rounded-lg hover:bg-viridian-700 text-sm"
                          >
                            ✏️
                          </button>
                          <button
                            onClick={() => handleDelete(book)}
                            className="px-3 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 text-sm"
                          >
                            🗑️
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};

export default BookList;
