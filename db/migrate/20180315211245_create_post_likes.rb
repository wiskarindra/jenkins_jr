class CreatePostLikes < ActiveRecord::Migration[5.1]
  def up
    create_table :post_likes do |t|
      t.integer :post_id
      t.integer :bukalapak_user_id
      t.boolean :liked, default: true
    end

    add_index :post_likes, [:post_id, :bukalapak_user_id], name: "index_post_likes_on_post_and_user", unique: true
    add_index :post_likes, [:post_id, :bukalapak_user_id, :liked], name: "index_post_likes_on_post_and_user_and_liked"
  end

  def down
    drop_table :post_likes
  end
end
