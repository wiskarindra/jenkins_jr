class AddReferenceToPostTags < ActiveRecord::Migration[5.1]
  def up
    add_column :post_tags, :reference_id, :bigint
    add_column :post_tags, :reference_type, :string
  end

  def down
    remove_column :post_tags, :reference_id
    remove_column :post_tags, :reference_type
  end
end
