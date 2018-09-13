class CreatePostTags < ActiveRecord::Migration[5.1]
  def up
    create_table :post_tags do |t|
      t.integer :post_id, index: true
      t.string  :name
      t.string  :url
      t.float   :coord_x
      t.float   :coord_y

      t.timestamps null: false
    end
  end

  def down
    drop_table :post_tags
  end
end
