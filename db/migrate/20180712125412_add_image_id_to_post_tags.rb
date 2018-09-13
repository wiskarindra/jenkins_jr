class AddImageIdToPostTags < ActiveRecord::Migration[5.1]
  def change
    add_column :post_tags, :post_image_id, :integer
  end
end
