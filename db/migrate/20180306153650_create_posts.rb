class CreatePosts < ActiveRecord::Migration[5.1]
  def up
    create_table :posts do |t|
      t.string   :title
      t.text     :description
      t.string   :influencer_name
      t.boolean  :published, default: false
      t.datetime :first_published_at
      t.datetime :last_published_at
      t.integer  :like_count, default: 0
      t.boolean  :deleted, default: false
      t.integer  :score, default: 0

      t.timestamps null: false
    end

    add_index :posts, [:deleted, :published], name: "index_posts_on_deleted_and_published"
  end

  def down
    drop_table :posts
  end
end
